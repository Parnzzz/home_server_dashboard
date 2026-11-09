using System.Globalization;
using System.Text.RegularExpressions;
using HomeServerDashboard.Api.Data;
using HomeServerDashboard.Api.Models;
using HomeServerDashboard.Api.Security;
using Microsoft.AspNetCore.Mvc;
using Npgsql;

namespace HomeServerDashboard.Api.Controllers;

[ApiController]
[Route("api/v1")]
public sealed class MetricsController(NpgsqlDataSource dataSource) : ControllerBase
{
    [HttpPost("metrics")]
    [ServiceFilter(typeof(AgentTokenFilter))]
    public async Task<IActionResult> Create(CreateMetricRequest request, CancellationToken cancellationToken)
    {
        if (!double.IsFinite(request.CpuPercent) || request.CpuPercent < 0 || request.CpuPercent > 100
            || request.MemoryUsedBytes < 0 || request.MemoryTotalBytes <= 0
            || request.DiskUsedBytes < 0 || request.DiskTotalBytes <= 0
            || request.MemoryUsedBytes > request.MemoryTotalBytes
            || request.DiskUsedBytes > request.DiskTotalBytes)
        {
            return BadRequest("invalid metric values");
        }

        if (HttpContext.Items["AgentServerId"] is not Guid serverId)
            return Unauthorized("Invalid or missing agent token");

        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        await MetricsRepository.CreateAsync(connection, serverId, request, cancellationToken);
        return StatusCode(StatusCodes.Status201Created);
    }

    [HttpGet("servers/{serverId:guid}/metrics")]
    [RequireUser]
    public async Task<IActionResult> GetByServer(
        Guid serverId,
        [FromQuery] string? limit,
        [FromQuery] string? from,
        CancellationToken cancellationToken)
    {
        var parsedLimit = 100;
        if (!string.IsNullOrWhiteSpace(limit)
            && (!int.TryParse(limit, NumberStyles.Integer, CultureInfo.InvariantCulture, out parsedLimit)))
        {
            return BadRequest("Invalid limit");
        }
        if (parsedLimit <= 0) parsedLimit = 100;
        parsedLimit = Math.Min(parsedLimit, 1000);

        DateTimeOffset? since = null;
        if (!string.IsNullOrWhiteSpace(from))
        {
            if (!GoDuration.TryParse(from, out var duration) || duration <= TimeSpan.Zero)
                return BadRequest("Invalid from parameter");
            since = DateTimeOffset.UtcNow.Subtract(duration);
        }

        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var server = await ServerRepository.GetOwnedByIdAsync(
            connection, serverId, HttpContext.GetUserId(), cancellationToken);
        if (server is null) return NotFound("Server not found");

        var metrics = await MetricsRepository.GetByServerAsync(
            connection, serverId, since, parsedLimit, cancellationToken);
        return Ok(metrics);
    }
}

internal static partial class GoDuration
{
    [GeneratedRegex(@"(?<value>[+-]?(?:\d+(?:\.\d*)?|\.\d+))(?<unit>ns|us|µs|μs|ms|s|m|h)", RegexOptions.CultureInvariant)]
    private static partial Regex TokenPattern();

    public static bool TryParse(string value, out TimeSpan result)
    {
        result = TimeSpan.Zero;
        var remaining = value.Trim();
        if (remaining.Length == 0) return false;
        var position = 0;
        double totalTicks = 0;
        foreach (Match match in TokenPattern().Matches(remaining))
        {
            if (match.Index != position
                || !double.TryParse(match.Groups["value"].Value, NumberStyles.Float,
                    CultureInfo.InvariantCulture, out var amount)) return false;
            var ticksPerUnit = match.Groups["unit"].Value switch
            {
                "ns" => TimeSpan.TicksPerSecond / 1_000_000_000d,
                "us" or "µs" or "μs" => TimeSpan.TicksPerMillisecond / 1000d,
                "ms" => TimeSpan.TicksPerMillisecond,
                "s" => TimeSpan.TicksPerSecond,
                "m" => TimeSpan.TicksPerMinute,
                "h" => TimeSpan.TicksPerHour,
                _ => 0,
            };
            totalTicks += amount * ticksPerUnit;
            position += match.Length;
        }

        if (position != remaining.Length || !double.IsFinite(totalTicks)
            || totalTicks <= 0 || totalTicks > TimeSpan.MaxValue.Ticks) return false;
        result = TimeSpan.FromTicks((long)totalTicks);
        return true;
    }
}
