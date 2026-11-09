using HomeServerDashboard.Api.Data;
using HomeServerDashboard.Api.Models;
using HomeServerDashboard.Api.Security;
using Microsoft.AspNetCore.Mvc;
using Npgsql;
using System.Text;

namespace HomeServerDashboard.Api.Controllers;

[ApiController]
[Route("api/v1/auth")]
public sealed class AuthController(
    NpgsqlDataSource dataSource,
    AppSettings settings,
    JwtTokenService tokenService) : ControllerBase
{
    [HttpPost("register")]
    public async Task<IActionResult> Register(RegisterRequest request, CancellationToken cancellationToken)
    {
        var email = (request.Email ?? "").Trim().ToLowerInvariant();
        var displayName = (request.DisplayName ?? "").Trim();
        var password = request.Password ?? "";
        if (email.Length == 0 || displayName.Length == 0 || password.Length == 0)
            return BadRequest(new { error = "email, display_name and password are required" });
        if (displayName.Length > 100 || Encoding.UTF8.GetByteCount(password) < 8
            || Encoding.UTF8.GetByteCount(password) > 72)
            return BadRequest(new { error = "invalid registration data" });

        var passwordHash = BCrypt.Net.BCrypt.HashPassword(password, workFactor: 10);
        try
        {
            await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
            await using var command = connection.CreateCommand();
            command.CommandText = """
                INSERT INTO users (email, display_name, password_hash)
                VALUES (@email, @display_name, @password_hash)
                RETURNING id, COALESCE(email, ''), display_name, avatar_url, role, password_hash, created_at
                """;
            command.Parameters.AddWithValue("email", email);
            command.Parameters.AddWithValue("display_name", displayName);
            command.Parameters.AddWithValue("password_hash", passwordHash);
            var user = await ReadUserAsync(command, cancellationToken);
            return StatusCode(StatusCodes.Status201Created, user.ToDto());
        }
        catch (PostgresException exception) when (exception.SqlState == PostgresErrorCodes.UniqueViolation)
        {
            return Conflict(new { error = "email already exists" });
        }
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login(LoginRequest request, CancellationToken cancellationToken)
    {
        var email = (request.Email ?? "").Trim();
        var password = request.Password ?? "";
        if (email.Length == 0 || password.Length == 0)
            return BadRequest(new { error = "email and password are required" });

        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var user = await UserRepository.FindByEmailAsync(connection, email, cancellationToken: cancellationToken);
        if (user?.PasswordHash is null || !BCrypt.Net.BCrypt.Verify(password, user.PasswordHash))
            return Unauthorized(new { error = "invalid email or password" });

        var response = await IssueTokensAsync(connection, user, cancellationToken);
        return Ok(response);
    }

    [HttpPost("refresh")]
    public async Task<IActionResult> Refresh(RefreshRequest request, CancellationToken cancellationToken)
    {
        var refreshToken = (request.RefreshToken ?? "").Trim();
        if (refreshToken.Length == 0)
            return BadRequest(new { error = "refresh_token is required" });

        var oldHash = JwtTokenService.HashToken(refreshToken);
        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        await using var transaction = await connection.BeginTransactionAsync(cancellationToken);

        await using var revoke = connection.CreateCommand();
        revoke.Transaction = transaction;
        revoke.CommandText = """
            UPDATE refresh_tokens
            SET revoked_at = NOW()
            WHERE token_hash = @token_hash
              AND revoked_at IS NULL
              AND expires_at > NOW()
            RETURNING user_id
            """;
        revoke.Parameters.AddWithValue("token_hash", oldHash);
        var userIdObject = await revoke.ExecuteScalarAsync(cancellationToken);
        if (userIdObject is not Guid userId)
            return Unauthorized(new { error = "invalid or expired refresh token" });

        var user = await UserRepository.FindByIdAsync(connection, userId, transaction, cancellationToken);
        if (user?.PasswordHash is null)
            return Unauthorized(new { error = "invalid or expired refresh token" });

        var newRefreshToken = JwtTokenService.CreateRandomToken();
        var accessToken = tokenService.CreateAccessToken(user.Id.ToString());
        await InsertRefreshTokenAsync(connection, transaction, user.Id, newRefreshToken, settings, cancellationToken);
        await transaction.CommitAsync(cancellationToken);

        return Ok(new AuthResponse(
            user.ToDto(), accessToken, newRefreshToken, "Bearer",
            (long)TimeSpan.FromMinutes(settings.AccessTokenMinutes).TotalSeconds));
    }

    [HttpPost("logout")]
    public async Task<IActionResult> Logout(RefreshRequest request, CancellationToken cancellationToken)
    {
        var refreshToken = (request.RefreshToken ?? "").Trim();
        if (refreshToken.Length == 0)
            return BadRequest(new { error = "refresh_token is required" });

        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        await using var command = connection.CreateCommand();
        command.CommandText = "UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = @token_hash AND revoked_at IS NULL";
        command.Parameters.AddWithValue("token_hash", JwtTokenService.HashToken(refreshToken));
        await command.ExecuteNonQueryAsync(cancellationToken);
        return Ok(new { message = "logged out" });
    }

    [HttpGet("me")]
    [RequireUser]
    public async Task<IActionResult> Me(CancellationToken cancellationToken)
    {
        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var user = await UserRepository.FindByIdAsync(connection, HttpContext.GetUserId(), cancellationToken: cancellationToken);
        return user is null ? Unauthorized(new { error = "unauthorized" }) : Ok(user.ToDto());
    }

    private async Task<AuthResponse> IssueTokensAsync(
        NpgsqlConnection connection,
        UserRecord user,
        CancellationToken cancellationToken)
    {
        var refreshToken = JwtTokenService.CreateRandomToken();
        var accessToken = tokenService.CreateAccessToken(user.Id.ToString());
        await InsertRefreshTokenAsync(connection, null, user.Id, refreshToken, settings, cancellationToken);
        return new AuthResponse(
            user.ToDto(), accessToken, refreshToken, "Bearer",
            (long)TimeSpan.FromMinutes(settings.AccessTokenMinutes).TotalSeconds);
    }

    private static async Task InsertRefreshTokenAsync(
        NpgsqlConnection connection,
        NpgsqlTransaction? transaction,
        Guid userId,
        string refreshToken,
        AppSettings settings,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.Transaction = transaction;
        command.CommandText = """
            INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
            VALUES (@user_id, @token_hash, @expires_at)
            """;
        command.Parameters.AddWithValue("user_id", userId);
        command.Parameters.AddWithValue("token_hash", JwtTokenService.HashToken(refreshToken));
        command.Parameters.AddWithValue("expires_at", DateTimeOffset.UtcNow.AddDays(settings.RefreshTokenDays));
        await command.ExecuteNonQueryAsync(cancellationToken);
    }

    private static async Task<UserRecord> ReadUserAsync(NpgsqlCommand command, CancellationToken cancellationToken)
    {
        await using var reader = await command.ExecuteReaderAsync(cancellationToken);
        if (!await reader.ReadAsync(cancellationToken)) throw new InvalidOperationException("User insert returned no row.");
        return UserRepository.ReadUser(reader);
    }
}
