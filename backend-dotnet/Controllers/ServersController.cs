using HomeServerDashboard.Api.Data;
using HomeServerDashboard.Api.Models;
using HomeServerDashboard.Api.Security;
using Microsoft.AspNetCore.Mvc;
using Npgsql;

namespace HomeServerDashboard.Api.Controllers;

[ApiController]
[RequireUser]
[Route("api/v1/servers")]
public sealed class ServersController(NpgsqlDataSource dataSource) : ControllerBase
{
    [HttpGet]
    public async Task<IActionResult> GetAll(CancellationToken cancellationToken)
    {
        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var servers = await ServerRepository.GetOwnedAsync(connection, HttpContext.GetUserId(), cancellationToken);
        return Ok(servers);
    }

    [HttpPost]
    public async Task<IActionResult> Create(CreateServerRequest request, CancellationToken cancellationToken)
    {
        var name = (request.Name ?? "").Trim();
        var hostname = (request.Hostname ?? "").Trim();
        var os = (request.Os ?? "").Trim();
        if (name.Length == 0 || name.Length > 100)
            return BadRequest("server name is required and must not exceed 100 characters");

        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var server = await ServerRepository.CreateAsync(
            connection, HttpContext.GetUserId(), name, hostname, os, cancellationToken);
        return StatusCode(StatusCodes.Status201Created, server);
    }

    [HttpGet("{serverId:guid}")]
    public async Task<IActionResult> GetById(Guid serverId, CancellationToken cancellationToken)
    {
        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var server = await ServerRepository.GetOwnedByIdAsync(
            connection, serverId, HttpContext.GetUserId(), cancellationToken);
        return server is null ? NotFound("Failed to retrieve server") : Ok(server);
    }

    [HttpDelete("{serverId:guid}")]
    public async Task<IActionResult> Delete(Guid serverId, CancellationToken cancellationToken)
    {
        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var deleted = await ServerRepository.DeleteOwnedAsync(
            connection, serverId, HttpContext.GetUserId(), cancellationToken);
        return deleted ? NoContent() : NotFound("Server not found");
    }

    [HttpPost("{serverId:guid}/agent-token")]
    public async Task<IActionResult> GenerateAgentToken(Guid serverId, CancellationToken cancellationToken)
    {
        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var server = await ServerRepository.GetOwnedByIdAsync(
            connection, serverId, HttpContext.GetUserId(), cancellationToken);
        if (server is null) return NotFound("Server not found");

        var token = JwtTokenService.CreateRandomToken();
        await using var command = connection.CreateCommand();
        command.CommandText = """
            INSERT INTO agent_tokens (server_id, token_hash)
            VALUES (@server_id, @token_hash)
            ON CONFLICT (server_id) DO UPDATE SET
                token_hash = EXCLUDED.token_hash,
                created_at = NOW(),
                last_used_at = NULL,
                revoked_at = NULL
            """;
        command.Parameters.AddWithValue("server_id", serverId);
        command.Parameters.AddWithValue("token_hash", JwtTokenService.HashToken(token));
        await command.ExecuteNonQueryAsync(cancellationToken);
        return StatusCode(StatusCodes.Status201Created, new { token, tokenType = "Bearer" });
    }

    [HttpDelete("{serverId:guid}/agent-token")]
    public async Task<IActionResult> RevokeAgentToken(Guid serverId, CancellationToken cancellationToken)
    {
        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        var server = await ServerRepository.GetOwnedByIdAsync(
            connection, serverId, HttpContext.GetUserId(), cancellationToken);
        if (server is null) return NotFound("Server not found");

        await using var command = connection.CreateCommand();
        command.CommandText = "UPDATE agent_tokens SET revoked_at = NOW() WHERE server_id = @server_id AND revoked_at IS NULL";
        command.Parameters.AddWithValue("server_id", serverId);
        await command.ExecuteNonQueryAsync(cancellationToken);
        return NoContent();
    }
}
