namespace HomeServerDashboard.Api.Models;

public sealed record UserDto(
    Guid Id,
    string Email,
    string DisplayName,
    string? AvatarUrl,
    string Role,
    DateTimeOffset CreatedAt);

public sealed record AuthResponse(
    UserDto User,
    string AccessToken,
    string RefreshToken,
    string TokenType,
    long ExpiresIn);

public sealed record ServerDto(
    Guid Id,
    string Name,
    string Hostname,
    string Os,
    string Status,
    DateTimeOffset CreatedAt);

public sealed record MetricDto(
    long Id,
    Guid ServerId,
    DateTimeOffset RecordedAt,
    double? CpuPercent,
    long? MemoryUsedBytes,
    long? MemoryTotalBytes,
    long? DiskUsedBytes,
    long? DiskTotalBytes);

public sealed record DashboardSummaryDto(
    ServerDto Server,
    double? CpuPercent,
    long? MemoryUsedBytes,
    long? MemoryTotalBytes,
    long? DiskUsedBytes,
    long? DiskTotalBytes,
    DateTimeOffset? LastSeenAt);

public sealed class RegisterRequest
{
    public string? Email { get; init; }
    public string? DisplayName { get; init; }
    public string? Password { get; init; }
}

public sealed class LoginRequest
{
    public string? Email { get; init; }
    public string? Password { get; init; }
}

public sealed class RefreshRequest
{
    public string? RefreshToken { get; init; }
}

public sealed class CreateServerRequest
{
    public string? Name { get; init; }
    public string? Hostname { get; init; }
    public string? Os { get; init; }
}

public sealed class CreateMetricRequest
{
    public double CpuPercent { get; init; }
    public long MemoryUsedBytes { get; init; }
    public long MemoryTotalBytes { get; init; }
    public long DiskUsedBytes { get; init; }
    public long DiskTotalBytes { get; init; }
}

public sealed class UpdateProfileRequest
{
    public string? DisplayName { get; init; }
    public string? AvatarUrl { get; init; }
}

public sealed class RequestPasswordResetRequest
{
    public string? Email { get; init; }
}

public sealed class ConfirmPasswordResetRequest
{
    public string? Email { get; init; }
    public string? Code { get; init; }
    public string? NewPassword { get; init; }
}
