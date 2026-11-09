using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;
using Npgsql;

namespace HomeServerDashboard.Api.Security;

[AttributeUsage(AttributeTargets.Class | AttributeTargets.Method)]
public sealed class RequireUserAttribute : Attribute, IAsyncAuthorizationFilter
{
    public Task OnAuthorizationAsync(AuthorizationFilterContext context)
    {
        var userId = context.HttpContext.User.FindFirst("user_id")?.Value;
        if (string.IsNullOrWhiteSpace(userId))
        {
            context.Result = new JsonResult(new { error = "unauthorized" })
            {
                StatusCode = StatusCodes.Status401Unauthorized,
            };
        }
        return Task.CompletedTask;
    }
}

public sealed class AgentTokenFilter(NpgsqlDataSource dataSource) : IAsyncAuthorizationFilter
{
    public async Task OnAuthorizationAsync(AuthorizationFilterContext context)
    {
        var header = context.HttpContext.Request.Headers["Authorization"].ToString();
        var parts = header.Split(' ', StringSplitOptions.RemoveEmptyEntries);
        if (parts.Length != 2 || !parts[0].Equals("Bearer", StringComparison.OrdinalIgnoreCase))
        {
            Unauthorized(context);
            return;
        }

        var hash = JwtTokenService.HashToken(parts[1]);
        await using var connection = await dataSource.OpenConnectionAsync(context.HttpContext.RequestAborted);
        await using var command = connection.CreateCommand();
        command.CommandText = """
            SELECT id, server_id
            FROM agent_tokens
            WHERE token_hash = @hash AND revoked_at IS NULL
            """;
        command.Parameters.AddWithValue("hash", hash);

        await using var reader = await command.ExecuteReaderAsync(context.HttpContext.RequestAborted);
        if (!await reader.ReadAsync(context.HttpContext.RequestAborted))
        {
            Unauthorized(context);
            return;
        }

        var tokenId = reader.GetGuid(0);
        var serverId = reader.GetGuid(1);
        await reader.CloseAsync();
        context.HttpContext.Items["AgentServerId"] = serverId;

        await using var update = connection.CreateCommand();
        update.CommandText = "UPDATE agent_tokens SET last_used_at = NOW() WHERE id = @id AND revoked_at IS NULL";
        update.Parameters.AddWithValue("id", tokenId);
        await update.ExecuteNonQueryAsync(context.HttpContext.RequestAborted);
    }

    private static void Unauthorized(AuthorizationFilterContext context)
    {
        context.Result = new JsonResult(new { error = "Invalid or missing agent token" })
        {
            StatusCode = StatusCodes.Status401Unauthorized,
        };
    }
}

public static class HttpContextIdentityExtensions
{
    public static Guid GetUserId(this HttpContext context)
    {
        var value = context.User.FindFirst("user_id")?.Value;
        return Guid.TryParse(value, out var id)
            ? id
            : throw new InvalidOperationException("Authenticated request is missing a valid user ID.");
    }
}
