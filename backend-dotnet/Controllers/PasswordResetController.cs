using HomeServerDashboard.Api.Models;
using HomeServerDashboard.Api.Services;
using Microsoft.AspNetCore.Mvc;

namespace HomeServerDashboard.Api.Controllers;

[ApiController]
[Route("api/v1/auth/password-reset")]
public sealed class PasswordResetController(PasswordResetService service) : ControllerBase
{
    private const string RequestMessage = "If an account exists for this email, a password reset code has been sent.";

    [HttpPost("request")]
    public async Task<IActionResult> RequestCode(
        RequestPasswordResetRequest request,
        CancellationToken cancellationToken)
    {
        if (string.IsNullOrWhiteSpace(request.Email))
            return BadRequest(new { error = "email is required" });

        var result = await service.RequestCodeAsync(request.Email, cancellationToken);
        return result switch
        {
            PasswordResetResult.InvalidEmail => BadRequest(new { error = "enter a valid email address" }),
            PasswordResetResult.Unavailable => StatusCode(StatusCodes.Status503ServiceUnavailable,
                new { error = "password reset email is not configured" }),
            PasswordResetResult.Accepted => Accepted(value: new { message = RequestMessage }),
            _ => StatusCode(StatusCodes.Status500InternalServerError,
                new { error = "password reset could not be completed" }),
        };
    }

    [HttpPost("confirm")]
    public async Task<IActionResult> Confirm(
        ConfirmPasswordResetRequest request,
        CancellationToken cancellationToken)
    {
        var result = await service.ResetPasswordAsync(
            request.Email ?? "", request.Code ?? "", request.NewPassword ?? "", cancellationToken);
        return result switch
        {
            PasswordResetResult.InvalidPassword => BadRequest(new { error = "password must be 8 to 72 bytes long" }),
            PasswordResetResult.InvalidCode or PasswordResetResult.InvalidEmail =>
                BadRequest(new { error = "invalid or expired reset code" }),
            PasswordResetResult.Accepted => Ok(new { message = "password reset successfully" }),
            _ => StatusCode(StatusCodes.Status500InternalServerError,
                new { error = "password reset could not be completed" }),
        };
    }
}
