using HomeServerDashboard.Api.Data;
using HomeServerDashboard.Api.Models;
using HomeServerDashboard.Api.Security;
using Microsoft.AspNetCore.Mvc;
using Npgsql;

namespace HomeServerDashboard.Api.Controllers;

[ApiController]
[RequireUser]
[Route("api/v1/servers")]
public sealed class DashboardController(NpgsqlDataSource dataSource) : ControllerBase
{
    [HttpGet("{serverId:guid}/summary")]
    public async Task<IActionResult> GetSummary(Guid serverId, CancellationToken cancellationToken)
    {
        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var server = await ServerRepository.GetOwnedByIdAsync(
            connection, serverId, HttpContext.GetUserId(), cancellationToken);
        if (server is null) return NotFound("Server not found");

        var latest = await MetricsRepository.GetLatestAsync(connection, serverId, cancellationToken);
        if (latest is null)
        {
            return Ok(new DashboardSummaryDto(
                server with { Status = "unknown" }, null, null, null, null, null, null));
        }

        var status = DateTimeOffset.UtcNow - latest.RecordedAt <= TimeSpan.FromSeconds(30)
            ? "online"
            : "offline";
        return Ok(new DashboardSummaryDto(
            server with { Status = status },
            latest.CpuPercent,
            latest.MemoryUsedBytes,
            latest.MemoryTotalBytes,
            latest.DiskUsedBytes,
            latest.DiskTotalBytes,
            latest.RecordedAt));
    }
}
