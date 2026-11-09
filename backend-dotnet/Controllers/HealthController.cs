using Microsoft.AspNetCore.Mvc;

namespace HomeServerDashboard.Api.Controllers;

[ApiController]
public sealed class HealthController : ControllerBase
{
    [HttpGet("/health")]
    public IActionResult GetHealth() => Ok(new
    {
        status = "ok",
        message = "Home Server Dashboard API is running",
        timestamp = DateTimeOffset.UtcNow,
    });
}
