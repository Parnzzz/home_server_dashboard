using System.Globalization;
using HomeServerDashboard.Api.Models;
using Npgsql;

namespace HomeServerDashboard.Api.Data;

public static class MetricsRepository
{
    private const string SelectColumns = """
        id, server_id, recorded_at, cpu_percent,
        memory_used_bytes, memory_total_bytes, disk_used_bytes, disk_total_bytes
        """;

    public static async Task<List<MetricDto>> GetByServerAsync(
        NpgsqlConnection connection,
        Guid serverId,
        DateTimeOffset? since,
        int limit,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.CommandText = since.HasValue
            ? $"SELECT {SelectColumns} FROM metrics WHERE server_id = @server_id AND recorded_at >= @since ORDER BY recorded_at DESC LIMIT @limit"
            : $"SELECT {SelectColumns} FROM metrics WHERE server_id = @server_id ORDER BY recorded_at DESC LIMIT @limit";
        command.Parameters.AddWithValue("server_id", serverId);
        command.Parameters.AddWithValue("limit", limit);
        if (since.HasValue) command.Parameters.AddWithValue("since", since.Value);

        var metrics = new List<MetricDto>();
        await using var reader = await command.ExecuteReaderAsync(cancellationToken);
        while (await reader.ReadAsync(cancellationToken)) metrics.Add(ReadMetric(reader));
        return metrics;
    }

    public static async Task<MetricDto?> GetLatestAsync(
        NpgsqlConnection connection,
        Guid serverId,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.CommandText = $"SELECT {SelectColumns} FROM metrics WHERE server_id = @server_id ORDER BY recorded_at DESC LIMIT 1";
        command.Parameters.AddWithValue("server_id", serverId);
        await using var reader = await command.ExecuteReaderAsync(cancellationToken);
        return await reader.ReadAsync(cancellationToken) ? ReadMetric(reader) : null;
    }

    public static async Task CreateAsync(
        NpgsqlConnection connection,
        Guid serverId,
        CreateMetricRequest request,
        CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();
        command.CommandText = """
            INSERT INTO metrics (
                server_id, cpu_percent, memory_used_bytes, memory_total_bytes,
                disk_used_bytes, disk_total_bytes
            ) VALUES (
                @server_id, @cpu, @memory_used, @memory_total, @disk_used, @disk_total
            )
            """;
        command.Parameters.AddWithValue("server_id", serverId);
        command.Parameters.AddWithValue("cpu", request.CpuPercent);
        command.Parameters.AddWithValue("memory_used", request.MemoryUsedBytes);
        command.Parameters.AddWithValue("memory_total", request.MemoryTotalBytes);
        command.Parameters.AddWithValue("disk_used", request.DiskUsedBytes);
        command.Parameters.AddWithValue("disk_total", request.DiskTotalBytes);
        await command.ExecuteNonQueryAsync(cancellationToken);
    }

    private static MetricDto ReadMetric(NpgsqlDataReader reader)
    {
        return new MetricDto(
            reader.GetInt64(0),
            reader.GetGuid(1),
            new DateTimeOffset(reader.GetDateTime(2)),
            reader.IsDBNull(3) ? null : Convert.ToDouble(reader.GetValue(3), CultureInfo.InvariantCulture),
            reader.IsDBNull(4) ? null : reader.GetInt64(4),
            reader.IsDBNull(5) ? null : reader.GetInt64(5),
            reader.IsDBNull(6) ? null : reader.GetInt64(6),
            reader.IsDBNull(7) ? null : reader.GetInt64(7));
    }
}
