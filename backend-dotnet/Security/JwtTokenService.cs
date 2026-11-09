using System.Security.Claims;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

namespace HomeServerDashboard.Api.Security;

public sealed class JwtTokenService(AppSettings settings)
{
    private const string Issuer = "home-server-dashboard";

    public string CreateAccessToken(string userId)
    {
        var now = DateTimeOffset.UtcNow;
        var header = Base64Url(JsonSerializer.SerializeToUtf8Bytes(new { alg = "HS256", typ = "JWT" }));
        var payload = Base64Url(JsonSerializer.SerializeToUtf8Bytes(new
        {
            user_id = userId,
            iss = Issuer,
            sub = userId,
            exp = now.AddMinutes(settings.AccessTokenMinutes).ToUnixTimeSeconds(),
            iat = now.ToUnixTimeSeconds(),
        }));
        var input = $"{header}.{payload}";
        using var hmac = new HMACSHA256(Encoding.UTF8.GetBytes(settings.JwtSecret));
        return $"{input}.{Base64Url(hmac.ComputeHash(Encoding.ASCII.GetBytes(input)))}";
    }

    public bool TryValidate(string token, out string userId)
    {
        userId = "";
        var parts = token.Split('.');
        if (parts.Length != 3) return false;

        try
        {
            using var headerDocument = JsonDocument.Parse(FromBase64Url(parts[0]));
            var header = headerDocument.RootElement;
            if (!header.TryGetProperty("alg", out var algorithm)
                || algorithm.GetString() != "HS256") return false;

            var input = $"{parts[0]}.{parts[1]}";
            using var hmac = new HMACSHA256(Encoding.UTF8.GetBytes(settings.JwtSecret));
            var expected = hmac.ComputeHash(Encoding.ASCII.GetBytes(input));
            var actual = FromBase64Url(parts[2]);
            if (actual.Length != expected.Length
                || !CryptographicOperations.FixedTimeEquals(actual, expected)) return false;

            using var payloadDocument = JsonDocument.Parse(FromBase64Url(parts[1]));
            var payload = payloadDocument.RootElement;
            if (!payload.TryGetProperty("iss", out var issuer)
                || issuer.GetString() != Issuer
                || !payload.TryGetProperty("exp", out var expires)
                || !payload.TryGetProperty("user_id", out var subject)) return false;

            if (expires.GetInt64() <= DateTimeOffset.UtcNow.ToUnixTimeSeconds()) return false;
            userId = subject.GetString() ?? "";
            return userId.Length > 0;
        }
        catch (Exception exception) when (
            exception is FormatException or JsonException or InvalidOperationException
                or ArgumentException or CryptographicException)
        {
            return false;
        }
    }

    public static string HashToken(string token)
    {
        return Convert.ToHexString(SHA256.HashData(Encoding.UTF8.GetBytes(token))).ToLowerInvariant();
    }

    public static string CreateRandomToken()
    {
        return Base64Url(RandomNumberGenerator.GetBytes(32));
    }

    private static string Base64Url(byte[] value)
    {
        return Convert.ToBase64String(value).TrimEnd('=').Replace('+', '-').Replace('/', '_');
    }

    private static byte[] FromBase64Url(string value)
    {
        var padded = value.Replace('-', '+').Replace('_', '/');
        padded = padded.PadRight(padded.Length + ((4 - padded.Length % 4) % 4), '=');
        return Convert.FromBase64String(padded);
    }
}

public sealed class JwtAuthenticationMiddleware(RequestDelegate next)
{
    public async Task InvokeAsync(HttpContext context, JwtTokenService tokenService)
    {
        var header = context.Request.Headers["Authorization"].ToString();
        var parts = header.Split(' ', StringSplitOptions.RemoveEmptyEntries);
        if (parts.Length == 2
            && parts[0].Equals("Bearer", StringComparison.OrdinalIgnoreCase)
            && tokenService.TryValidate(parts[1], out var userId))
        {
            context.User = new ClaimsPrincipal(new ClaimsIdentity(
                [new Claim("user_id", userId), new Claim(ClaimTypes.NameIdentifier, userId)],
                "Bearer"));
        }

        await next(context);
    }
}
