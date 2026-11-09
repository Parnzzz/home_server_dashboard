using System.Text.Json;
using System.Text.Json.Serialization;
using HomeServerDashboard.Api;
using HomeServerDashboard.Api.Security;
using HomeServerDashboard.Api.Services;
using Npgsql;

var builder = WebApplication.CreateBuilder(args);
var settings = AppSettings.Load(builder.Configuration);

builder.WebHost.UseUrls(
    Environment.GetEnvironmentVariable("ASPNETCORE_URLS")
        ?? $"http://0.0.0.0:{settings.Port}");

builder.Services.AddSingleton(settings);
builder.Services.AddSingleton(_ => NpgsqlDataSource.Create(settings.DatabaseConnectionString));
builder.Services.AddSingleton<JwtTokenService>();
builder.Services.AddScoped<AgentTokenFilter>();
builder.Services.AddScoped<PasswordResetService>();
builder.Services.AddControllers().AddJsonOptions(options =>
{
    options.JsonSerializerOptions.PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower;
    options.JsonSerializerOptions.DictionaryKeyPolicy = JsonNamingPolicy.SnakeCaseLower;
    options.JsonSerializerOptions.DefaultIgnoreCondition = JsonIgnoreCondition.Never;
});

builder.Services.AddSingleton<IPasswordResetMailer>(services =>
{
    if (settings.SmtpHost.Length > 0) return new SmtpPasswordResetMailer(settings);
    if (settings.Environment == "production") return new UnavailablePasswordResetMailer();
    return new ConsolePasswordResetMailer(services.GetRequiredService<ILogger<ConsolePasswordResetMailer>>());
});
builder.Services.AddHostedService<RetentionCleanupService>();

var app = builder.Build();
await using (var startupConnection = await app.Services
    .GetRequiredService<NpgsqlDataSource>()
    .OpenConnectionAsync())
{
    app.Logger.LogInformation("Connected to PostgreSQL successfully");
}

app.UseRouting();
app.UseMiddleware<JwtAuthenticationMiddleware>();
app.MapControllers();

app.Logger.LogInformation("Home Server Dashboard ASP.NET API starting on port {Port}", settings.Port);
app.Run();

//$env:DATABASE_URL = "ใส่ connection string ของ PostgreSQL"
//$env:JWT_SECRET = "ใส่ JWT_SECRET เดียวกับ Go backend"
//$env:APP_ENV = "development"
//$env:PORT = "8081"
//dotnet run
