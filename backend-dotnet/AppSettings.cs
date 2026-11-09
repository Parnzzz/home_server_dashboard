using System.Net.Mail;
using System.Text;
using Npgsql;

namespace HomeServerDashboard.Api;

public sealed class AppSettings
{
    public required string Environment { get; init; }
    public required string DatabaseConnectionString { get; init; }
    public required string JwtSecret { get; init; }
    public required int Port { get; init; }
    public required int AccessTokenMinutes { get; init; }
    public required int RefreshTokenDays { get; init; }
    public string SmtpHost { get; init; } = "";
    public int SmtpPort { get; init; } = 587;
    public string SmtpUsername { get; init; } = "";
    public string SmtpPassword { get; init; } = "";
    public string SmtpFrom { get; init; } = "";

    public static AppSettings Load(IConfiguration configuration)
    {
        var environment = (System.Environment.GetEnvironmentVariable("APP_ENV") ?? "development")
            .Trim().ToLowerInvariant();
        if (environment is not ("development" or "testing" or "production"))
        {
            throw new InvalidOperationException("APP_ENV must be development, testing, or production.");
        }

        var databaseUrl = System.Environment.GetEnvironmentVariable("DATABASE_URL")
            ?? configuration.GetConnectionString("Database");
        if (string.IsNullOrWhiteSpace(databaseUrl))
        {
            throw new InvalidOperationException("DATABASE_URL is required.");
        }

        var jwtSecret = System.Environment.GetEnvironmentVariable("JWT_SECRET") ?? "";
        if (Encoding.UTF8.GetByteCount(jwtSecret) < 32)
        {
            throw new InvalidOperationException("JWT_SECRET must be at least 32 bytes.");
        }

        var smtpHost = (System.Environment.GetEnvironmentVariable("SMTP_HOST") ?? "").Trim();
        var smtpUsername = System.Environment.GetEnvironmentVariable("SMTP_USERNAME") ?? "";
        var smtpPassword = System.Environment.GetEnvironmentVariable("SMTP_PASSWORD") ?? "";
        var smtpFrom = (System.Environment.GetEnvironmentVariable("SMTP_FROM") ?? "").Trim();
        var smtpPort = ReadInt("SMTP_PORT", 587, 1, 65535);

        if (smtpHost.Length > 0)
        {
            if (smtpFrom.Length == 0 || !MailAddress.TryCreate(smtpFrom, out var address)
                || address.Address != smtpFrom)
            {
                throw new InvalidOperationException("SMTP_FROM must be a valid email when SMTP_HOST is configured.");
            }
        }
        else if (smtpUsername.Length > 0 || smtpPassword.Length > 0 || smtpFrom.Length > 0)
        {
            throw new InvalidOperationException("SMTP_HOST is required when SMTP settings are configured.");
        }

        if ((smtpUsername.Length == 0) != (smtpPassword.Length == 0))
        {
            throw new InvalidOperationException("SMTP_USERNAME and SMTP_PASSWORD must be set together.");
        }

        return new AppSettings
        {
            Environment = environment,
            DatabaseConnectionString = ToNpgsqlConnectionString(databaseUrl),
            JwtSecret = jwtSecret,
            Port = ReadInt("PORT", 8080, 1, 65535),
            AccessTokenMinutes = ReadInt("ACCESS_TOKEN_MINUTES", 15, 1, int.MaxValue),
            RefreshTokenDays = ReadInt("REFRESH_TOKEN_DAYS", 7, 1, int.MaxValue),
            SmtpHost = smtpHost,
            SmtpPort = smtpPort,
            SmtpUsername = smtpUsername,
            SmtpPassword = smtpPassword,
            SmtpFrom = smtpFrom,
        };
    }

    private static int ReadInt(string name, int defaultValue, int min, int max)
    {
        var value = System.Environment.GetEnvironmentVariable(name);
        if (string.IsNullOrWhiteSpace(value)) return defaultValue;
        if (!int.TryParse(value, out var parsed) || parsed < min || parsed > max)
        {
            throw new InvalidOperationException($"{name} has an invalid value.");
        }
        return parsed;
    }

    private static string ToNpgsqlConnectionString(string value)
    {
        if (!Uri.TryCreate(value, UriKind.Absolute, out var uri)
            || (uri.Scheme != "postgres" && uri.Scheme != "postgresql"))
        {
            return value;
        }

        var userInfo = uri.UserInfo.Split(':', 2);
        var builder = new NpgsqlConnectionStringBuilder
        {
            Host = uri.Host,
            Port = uri.IsDefaultPort ? 5432 : uri.Port,
            Database = Uri.UnescapeDataString(uri.AbsolutePath.Trim('/')),
            Username = userInfo.Length > 0 ? Uri.UnescapeDataString(userInfo[0]) : "",
            Password = userInfo.Length > 1 ? Uri.UnescapeDataString(userInfo[1]) : "",
        };

        foreach (var pair in uri.Query.TrimStart('?').Split('&', StringSplitOptions.RemoveEmptyEntries))
        {
            var parts = pair.Split('=', 2);
            if (parts.Length != 2) continue;
            var key = Uri.UnescapeDataString(parts[0]);
            var setting = Uri.UnescapeDataString(parts[1]);
            if (key.Equals("sslmode", StringComparison.OrdinalIgnoreCase))
            {
                builder.SslMode = setting.ToLowerInvariant() switch
                {
                    "disable" => SslMode.Disable,
                    "allow" => SslMode.Allow,
                    "prefer" => SslMode.Prefer,
                    "require" => SslMode.Require,
                    "verify-ca" => SslMode.VerifyCA,
                    "verify-full" => SslMode.VerifyFull,
                    _ => throw new InvalidOperationException("DATABASE_URL has an invalid sslmode."),
                };
            }
        }

        return builder.ConnectionString;
    }
}
