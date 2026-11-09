using System.Net.Mail;
using System.Globalization;
using System.Security.Cryptography;
using System.Text;
using HomeServerDashboard.Api.Data;
using Npgsql;

namespace HomeServerDashboard.Api.Services;

public enum PasswordResetResult
{
    Accepted,
    Unavailable,
    InvalidEmail,
    InvalidCode,
    InvalidPassword,
    Failed,
}

public sealed class PasswordResetService(
    NpgsqlDataSource dataSource,
    AppSettings settings,
    IPasswordResetMailer mailer,
    ILogger<PasswordResetService> logger)
{
    private const int MaximumAttempts = 5;

    public async Task<PasswordResetResult> RequestCodeAsync(string email, CancellationToken cancellationToken)
    {
        if (!mailer.Available) return PasswordResetResult.Unavailable;
        if (!TryNormalizeEmail(email, out email)) return PasswordResetResult.InvalidEmail;

        try
        {
            await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
            var user = await UserRepository.FindByEmailAsync(connection, email, cancellationToken: cancellationToken);
            if (user is null) return PasswordResetResult.Accepted;

            var code = RandomNumberGenerator.GetInt32(100_000_000).ToString("D8", CultureInfo.InvariantCulture);
            var codeHash = HashCode(user.Id, code);
            var created = await CreateCodeAsync(connection, user.Id, codeHash, cancellationToken);
            if (!created) return PasswordResetResult.Accepted;

            try
            {
                await mailer.SendCodeAsync(email, code, cancellationToken);
            }
            catch (Exception exception)
            {
                logger.LogError(exception, "Password reset email could not be sent.");
                await InvalidateLatestCodeAsync(connection, user.Id, cancellationToken);
                return PasswordResetResult.Accepted;
            }

            return PasswordResetResult.Accepted;
        }
        catch (Exception exception) when (exception is not OperationCanceledException)
        {
            logger.LogError(exception, "Password reset request could not be completed.");
            return PasswordResetResult.Accepted;
        }
    }

    public async Task<PasswordResetResult> ResetPasswordAsync(
        string email,
        string code,
        string newPassword,
        CancellationToken cancellationToken)
    {
        if (!TryNormalizeEmail(email, out email) || !IsValidCode(code))
            return PasswordResetResult.InvalidCode;
        var passwordBytes = Encoding.UTF8.GetByteCount(newPassword);
        if (passwordBytes < 8 || passwordBytes > 72)
            return PasswordResetResult.InvalidPassword;

        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var user = await UserRepository.FindByEmailAsync(connection, email, cancellationToken: cancellationToken);
        if (user is null) return PasswordResetResult.InvalidCode;

        await using var lookup = connection.CreateCommand();
        lookup.CommandText = """
            SELECT id, code_hash, expires_at, attempts, used_at
            FROM password_reset_codes
            WHERE user_id = @user_id AND used_at IS NULL
            ORDER BY created_at DESC
            LIMIT 1
            """;
        lookup.Parameters.AddWithValue("user_id", user.Id);
        Guid codeId;
        string storedHash;
        DateTimeOffset expiresAt;
        short attempts;
        DateTimeOffset? usedAt;
        await using (var reader = await lookup.ExecuteReaderAsync(cancellationToken))
        {
            if (!await reader.ReadAsync(cancellationToken)) return PasswordResetResult.InvalidCode;
            codeId = reader.GetGuid(0);
            storedHash = reader.GetString(1).TrimEnd();
            expiresAt = new DateTimeOffset(reader.GetDateTime(2));
            attempts = reader.GetInt16(3);
            usedAt = reader.IsDBNull(4) ? null : new DateTimeOffset(reader.GetDateTime(4));
        }

        if (usedAt.HasValue || attempts >= MaximumAttempts || expiresAt <= DateTimeOffset.UtcNow)
            return PasswordResetResult.InvalidCode;

        var expectedHash = HashCode(user.Id, code);
        if (!CryptographicOperations.FixedTimeEquals(
                Encoding.ASCII.GetBytes(storedHash), Encoding.ASCII.GetBytes(expectedHash)))
        {
            await IncrementFailedAttemptAsync(connection, codeId, cancellationToken);
            return PasswordResetResult.InvalidCode;
        }

        var passwordHash = BCrypt.Net.BCrypt.HashPassword(newPassword, workFactor: 10);
        await using var transaction = await connection.BeginTransactionAsync(cancellationToken);
        await using var lockCode = connection.CreateCommand();
        lockCode.Transaction = transaction;
        lockCode.CommandText = """
            SELECT code_hash, expires_at, attempts, used_at
            FROM password_reset_codes
            WHERE id = @id AND user_id = @user_id
            FOR UPDATE
            """;
        lockCode.Parameters.AddWithValue("id", codeId);
        lockCode.Parameters.AddWithValue("user_id", user.Id);
        string currentHash;
        DateTimeOffset currentExpiry;
        short currentAttempts;
        DateTimeOffset? currentUsed;
        await using (var reader = await lockCode.ExecuteReaderAsync(cancellationToken))
        {
            if (!await reader.ReadAsync(cancellationToken)) return PasswordResetResult.InvalidCode;
            currentHash = reader.GetString(0).TrimEnd();
            currentExpiry = new DateTimeOffset(reader.GetDateTime(1));
            currentAttempts = reader.GetInt16(2);
            currentUsed = reader.IsDBNull(3) ? null : new DateTimeOffset(reader.GetDateTime(3));
        }

        if (currentUsed.HasValue || currentAttempts >= MaximumAttempts
            || currentExpiry <= DateTimeOffset.UtcNow
            || !CryptographicOperations.FixedTimeEquals(
                Encoding.ASCII.GetBytes(currentHash), Encoding.ASCII.GetBytes(expectedHash)))
        {
            return PasswordResetResult.InvalidCode;
        }

        await using var updateUser = connection.CreateCommand();
        updateUser.Transaction = transaction;
        updateUser.CommandText = "UPDATE users SET password_hash = @password_hash WHERE id = @user_id";
        updateUser.Parameters.AddWithValue("password_hash", passwordHash);
        updateUser.Parameters.AddWithValue("user_id", user.Id);
        if (await updateUser.ExecuteNonQueryAsync(cancellationToken) != 1)
            return PasswordResetResult.InvalidCode;

        await using var consume = connection.CreateCommand();
        consume.Transaction = transaction;
        consume.CommandText = "UPDATE password_reset_codes SET used_at = NOW() WHERE id = @id";
        consume.Parameters.AddWithValue("id", codeId);
        await consume.ExecuteNonQueryAsync(cancellationToken);

        await using var revoke = connection.CreateCommand();
        revoke.Transaction = transaction;
        revoke.CommandText = "UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = @user_id AND revoked_at IS NULL";
        revoke.Parameters.AddWithValue("user_id", user.Id);
        await revoke.ExecuteNonQueryAsync(cancellationToken);

        await transaction.CommitAsync(cancellationToken);
        return PasswordResetResult.Accepted;
    }

    private async Task<bool> CreateCodeAsync(
        NpgsqlConnection connection,
        Guid userId,
        string codeHash,
        CancellationToken cancellationToken)
    {
        await using var transaction = await connection.BeginTransactionAsync(cancellationToken);
        await using var lockUser = connection.CreateCommand();
        lockUser.Transaction = transaction;
        lockUser.CommandText = "SELECT id FROM users WHERE id = @id FOR UPDATE";
        lockUser.Parameters.AddWithValue("id", userId);
        if (await lockUser.ExecuteScalarAsync(cancellationToken) is null) return false;

        await using var cooldown = connection.CreateCommand();
        cooldown.Transaction = transaction;
        cooldown.CommandText = """
            SELECT EXISTS (
                SELECT 1 FROM password_reset_codes
                WHERE user_id = @user_id AND created_at > NOW() - INTERVAL '1 minute'
            )
            """;
        cooldown.Parameters.AddWithValue("user_id", userId);
        if (Convert.ToBoolean(await cooldown.ExecuteScalarAsync(cancellationToken), CultureInfo.InvariantCulture))
            return false;

        await using var invalidate = connection.CreateCommand();
        invalidate.Transaction = transaction;
        invalidate.CommandText = "UPDATE password_reset_codes SET used_at = NOW() WHERE user_id = @user_id AND used_at IS NULL";
        invalidate.Parameters.AddWithValue("user_id", userId);
        await invalidate.ExecuteNonQueryAsync(cancellationToken);

        await using var insert = connection.CreateCommand();
        insert.Transaction = transaction;
        insert.CommandText = """
            INSERT INTO password_reset_codes (user_id, code_hash, expires_at)
            VALUES (@user_id, @code_hash, NOW() + INTERVAL '15 minutes')
            """;
        insert.Parameters.AddWithValue("user_id", userId);
        insert.Parameters.AddWithValue("code_hash", codeHash);
        await insert.ExecuteNonQueryAsync(cancellationToken);
        await transaction.CommitAsync(cancellationToken);
        return true;
    }

    private static async Task IncrementFailedAttemptAsync(
        NpgsqlConnection connection,
        Guid codeId,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.CommandText = """
            UPDATE password_reset_codes
            SET attempts = attempts + 1,
                used_at = CASE WHEN attempts + 1 >= 5 THEN NOW() ELSE used_at END
            WHERE id = @id AND used_at IS NULL AND attempts < 5
            """;
        command.Parameters.AddWithValue("id", codeId);
        await command.ExecuteNonQueryAsync(cancellationToken);
    }

    private static async Task InvalidateLatestCodeAsync(
        NpgsqlConnection connection,
        Guid userId,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.CommandText = """
            UPDATE password_reset_codes SET used_at = NOW()
            WHERE id = (
                SELECT id FROM password_reset_codes
                WHERE user_id = @user_id AND used_at IS NULL
                ORDER BY created_at DESC LIMIT 1
            )
            """;
        command.Parameters.AddWithValue("user_id", userId);
        await command.ExecuteNonQueryAsync(cancellationToken);
    }

    private string HashCode(Guid userId, string code)
    {
        var data = Encoding.UTF8.GetBytes($"{userId:D}:{code}");
        var hash = HMACSHA256.HashData(Encoding.UTF8.GetBytes(settings.JwtSecret), data);
        return Convert.ToHexString(hash).ToLowerInvariant();
    }

    private static bool IsValidCode(string code)
    {
        return code.Length == 8 && code.All(char.IsAsciiDigit);
    }

    private static bool TryNormalizeEmail(string? value, out string email)
    {
        email = (value ?? "").Trim().ToLowerInvariant();
        if (!MailAddress.TryCreate(email, out var address) || address.Address != email)
            return false;
        return true;
    }
}
