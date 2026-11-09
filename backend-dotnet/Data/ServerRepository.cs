using HomeServerDashboard.Api.Models;
using Npgsql;

namespace HomeServerDashboard.Api.Data;

public static class ServerRepository
{
    public static async Task<List<ServerDto>> GetOwnedAsync(
        NpgsqlConnection connection,
        Guid userId,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.CommandText = """
            SELECT id, name, COALESCE(hostname, ''), COALESCE(os, ''), status, created_at
            FROM servers
            WHERE user_id = @user_id
            ORDER BY created_at DESC
            """;
        command.Parameters.AddWithValue("user_id", userId);

        var servers = new List<ServerDto>();
        await using var reader = await command.ExecuteReaderAsync(cancellationToken);
        while (await reader.ReadAsync(cancellationToken)) servers.Add(ReadServer(reader));
        return servers;
    }

    public static async Task<ServerDto?> GetOwnedByIdAsync(
        NpgsqlConnection connection,
        Guid serverId,
        Guid userId,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.CommandText = """
            SELECT id, name, COALESCE(hostname, ''), COALESCE(os, ''), status, created_at
            FROM servers
            WHERE id = @id AND user_id = @user_id
            """;
        command.Parameters.AddWithValue("id", serverId);
        command.Parameters.AddWithValue("user_id", userId);
        await using var reader = await command.ExecuteReaderAsync(cancellationToken);
        return await reader.ReadAsync(cancellationToken) ? ReadServer(reader) : null;
    }

    public static async Task<ServerDto> CreateAsync(
        NpgsqlConnection connection,
        Guid userId,
        string name,
        string hostname,
        string os,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.CommandText = """
            INSERT INTO servers (user_id, name, hostname, os)
            VALUES (@user_id, @name, NULLIF(@hostname, ''), NULLIF(@os, ''))
            RETURNING id, name, COALESCE(hostname, ''), COALESCE(os, ''), status, created_at
            """;
        command.Parameters.AddWithValue("user_id", userId);
        command.Parameters.AddWithValue("name", name);
        command.Parameters.AddWithValue("hostname", hostname);
        command.Parameters.AddWithValue("os", os);
        await using var reader = await command.ExecuteReaderAsync(cancellationToken);
        if (!await reader.ReadAsync(cancellationToken)) throw new InvalidOperationException("Server insert returned no row.");
        return ReadServer(reader);
    }

    public static async Task<bool> DeleteOwnedAsync(
        NpgsqlConnection connection,
        Guid serverId,
        Guid userId,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.CommandText = "DELETE FROM servers WHERE id = @id AND user_id = @user_id";
        command.Parameters.AddWithValue("id", serverId);
        command.Parameters.AddWithValue("user_id", userId);
        return await command.ExecuteNonQueryAsync(cancellationToken) > 0;
    }

    public static ServerDto ReadServer(NpgsqlDataReader reader)
    {
        return new ServerDto(
            reader.GetGuid(0), reader.GetString(1), reader.GetString(2), reader.GetString(3),
            reader.GetString(4), new DateTimeOffset(reader.GetDateTime(5)));
    }
}
