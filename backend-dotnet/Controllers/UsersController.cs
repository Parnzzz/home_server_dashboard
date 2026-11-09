using HomeServerDashboard.Api.Data;
using HomeServerDashboard.Api.Models;
using HomeServerDashboard.Api.Security;
using Microsoft.AspNetCore.Mvc;
using Npgsql;

namespace HomeServerDashboard.Api.Controllers;

[ApiController]
[RequireUser]
[Route("api/v1/users")]
public sealed class UsersController(NpgsqlDataSource dataSource) : ControllerBase
{
    [HttpPatch("me")]
    public async Task<IActionResult> UpdateProfile(UpdateProfileRequest request, CancellationToken cancellationToken)
    {
        var displayName = (request.DisplayName ?? "").Trim();
        var avatarUrl = request.AvatarUrl?.Trim();
        if (displayName.Length == 0 || displayName.Length > 100 || avatarUrl == "")
            return BadRequest(new { error = "invalid profile data" });

        await using var connection = await dataSource.OpenConnectionAsync(cancellationToken);
        await using var command = connection.CreateCommand();
        command.CommandText = """
            UPDATE users
            SET display_name = @display_name,
                avatar_url = COALESCE(@avatar_url, avatar_url)
            WHERE id = @user_id
            RETURNING id, COALESCE(email, ''), display_name, avatar_url, role, password_hash, created_at
            """;
        command.Parameters.AddWithValue("user_id", HttpContext.GetUserId());
        command.Parameters.AddWithValue("display_name", displayName);
        command.Parameters.Add(new NpgsqlParameter("avatar_url", NpgsqlTypes.NpgsqlDbType.Text)
        {
            Value = (object?)avatarUrl ?? DBNull.Value,
        });

        await using var reader = await command.ExecuteReaderAsync(cancellationToken);
        if (!await reader.ReadAsync(cancellationToken))
            return Unauthorized(new { error = "unauthorized" });
        return Ok(UserRepository.ReadUser(reader).ToDto());
    }
}
