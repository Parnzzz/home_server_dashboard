using MailKit.Net.Smtp;
using MailKit.Security;
using MimeKit;

namespace HomeServerDashboard.Api.Services;

public interface IPasswordResetMailer
{
    bool Available { get; }
    Task SendCodeAsync(string email, string code, CancellationToken cancellationToken);
}

public sealed class SmtpPasswordResetMailer(AppSettings settings) : IPasswordResetMailer
{
    public bool Available => settings.SmtpHost.Length > 0 && settings.SmtpFrom.Length > 0;

    public async Task SendCodeAsync(string email, string code, CancellationToken cancellationToken)
    {
        var message = new MimeMessage();
        message.From.Add(MailboxAddress.Parse(settings.SmtpFrom));
        message.To.Add(MailboxAddress.Parse(email));
        message.Subject = "Password reset code";
        message.Body = new TextPart("plain")
        {
            Text = $"Your Home Server Dashboard password reset code is {code}.\r\n"
                + "It expires in 15 minutes. If you did not request this, you can ignore this email.\r\n",
        };

        using var client = new SmtpClient();
        var socketOptions = settings.SmtpPort == 465
            ? SecureSocketOptions.SslOnConnect
            : SecureSocketOptions.StartTls;
        await client.ConnectAsync(settings.SmtpHost, settings.SmtpPort, socketOptions, cancellationToken);
        if (settings.SmtpUsername.Length > 0)
            await client.AuthenticateAsync(settings.SmtpUsername, settings.SmtpPassword, cancellationToken);
        await client.SendAsync(message, cancellationToken);
        await client.DisconnectAsync(true, cancellationToken);
    }
}

public sealed class ConsolePasswordResetMailer(ILogger<ConsolePasswordResetMailer> logger) : IPasswordResetMailer
{
    public bool Available => true;

    public Task SendCodeAsync(string email, string code, CancellationToken cancellationToken)
    {
        logger.LogInformation("[development only] password reset email to {Email}; code: {Code}", email, code);
        return Task.CompletedTask;
    }
}

public sealed class UnavailablePasswordResetMailer : IPasswordResetMailer
{
    public bool Available => false;

    public Task SendCodeAsync(string email, string code, CancellationToken cancellationToken)
    {
        throw new InvalidOperationException("password reset email is not configured");
    }
}
