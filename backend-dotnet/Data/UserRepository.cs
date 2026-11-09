using HomeServerDashboard.Api.Models;
using Npgsql;

namespace HomeServerDashboard.Api.Data;

public sealed record UserRecord(
    Guid Id,
    string Email,
    string DisplayName,
    string? AvatarUrl,
    string Role,
    string? PasswordHash,
    DateTimeOffset CreatedAt)
{
    public UserDto ToDto() => new(Id, Email, DisplayName, AvatarUrl, Role, CreatedAt);
}

public static class UserRepository
{
    public static async Task<UserRecord?> FindByEmailAsync(
        NpgsqlConnection connection,
        string email,
        NpgsqlTransaction? transaction = null,
        CancellationToken cancellationToken = default)
    {
        await using var command = connection.CreateCommand();
        command.Transaction = transaction;
        command.CommandText = """
            SELECT id, COALESCE(email, ''), display_name, avatar_url, role, password_hash, created_at
            FROM users
            WHERE LOWER(email) = LOWER(@email)
            """;
        command.Parameters.AddWithValue("email", email);
        await using var reader = await command.ExecuteReaderAsync(cancellationToken);
        return await reader.ReadAsync(cancellationToken)
            ? ReadUser(reader)
            : null;
    }

    public static async Task<UserRecord?> FindByIdAsync(
        NpgsqlConnection connection,
        Guid userId,
        NpgsqlTransaction? transaction = null,
        CancellationToken cancellationToken = default)
    {
        await using var command = connection.CreateCommand();
        command.Transaction = transaction;
        command.CommandText = """
            SELECT id, COALESCE(email, ''), display_name, avatar_url, role, password_hash, created_at
            FROM users
            WHERE id = @id
            """;
        command.Parameters.AddWithValue("id", userId);
        await using var reader = await command.ExecuteReaderAsync(cancellationToken);
        return await reader.ReadAsync(cancellationToken)
            ? ReadUser(reader)
            : null;
    }

    public static UserRecord ReadUser(NpgsqlDataReader reader)
    {
        return new UserRecord(
            reader.GetGuid(0),
            reader.GetString(1),
            reader.GetString(2),
            reader.IsDBNull(3) ? null : reader.GetString(3),
            reader.GetString(4),
            reader.IsDBNull(5) ? null : reader.GetString(5),
            new DateTimeOffset(reader.GetDateTime(6)));
    }
}
