using Npgsql;

namespace HomeServerDashboard.Api.Services;

public sealed class RetentionCleanupService(
    NpgsqlDataSource dataSource,
    ILogger<RetentionCleanupService> logger) : BackgroundService
{
    private const int BatchSize = 2000;
    private static readonly TimeSpan CleanupInterval = TimeSpan.FromHours(24);

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        using var timer = new PeriodicTimer(CleanupInterval);
        do
        {
            await CleanupAsync(stoppingToken);
        }
        while (await timer.WaitForNextTickAsync(stoppingToken));
    }

    private async Task CleanupAsync(CancellationToken stoppingToken)
    {
        using var timeout = CancellationTokenSource.CreateLinkedTokenSource(stoppingToken);
        timeout.CancelAfter(TimeSpan.FromMinutes(30));

        try
        {
            await using var connection = await dataSource.OpenConnectionAsync(timeout.Token);
            await using var schema = connection.CreateCommand();
            schema.CommandText = """
                SELECT
                    to_regclass('metrics') IS NOT NULL
                    AND to_regclass('refresh_tokens') IS NOT NULL
                    AND to_regclass('password_reset_codes') IS NOT NULL
                    AND to_regclass('idx_metrics_recorded_at') IS NOT NULL
                    AND to_regclass('idx_refresh_tokens_expires_at') IS NOT NULL
                    AND to_regclass('idx_refresh_tokens_revoked_at') IS NOT NULL
                    AND to_regclass('idx_password_reset_codes_created_at') IS NOT NULL
                """;
            var ready = Convert.ToBoolean(await schema.ExecuteScalarAsync(timeout.Token));
            if (!ready)
            {
                logger.LogWarning("Database retention cleanup skipped; apply migrations 000005 and 000006 first.");
                return;
            }

            var metrics = await DeleteInBatchesAsync(connection, """
                WITH expired AS (
                    SELECT id FROM metrics
                    WHERE recorded_at < NOW() - INTERVAL '7 days'
                    ORDER BY recorded_at LIMIT @batch_size
                )
                DELETE FROM metrics AS metric USING expired WHERE metric.id = expired.id
                """, timeout.Token);

            var refreshTokens = await DeleteInBatchesAsync(connection, """
                WITH expired AS (
                    SELECT id FROM refresh_tokens
                    WHERE (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '30 days')
                       OR expires_at < NOW() - INTERVAL '30 days'
                    ORDER BY created_at LIMIT @batch_size
                )
                DELETE FROM refresh_tokens AS token USING expired WHERE token.id = expired.id
                """, timeout.Token);

            var resetCodes = await DeleteInBatchesAsync(connection, """
                WITH expired AS (
                    SELECT id FROM password_reset_codes
                    WHERE created_at < NOW() - INTERVAL '1 day'
                    ORDER BY created_at LIMIT @batch_size
                )
                DELETE FROM password_reset_codes AS code USING expired WHERE code.id = expired.id
                """, timeout.Token);

            logger.LogInformation(
                "Database retention cleanup completed: metrics={Metrics}, refresh_tokens={RefreshTokens}, password_reset_codes={PasswordResetCodes}",
                metrics, refreshTokens, resetCodes);
        }
        catch (OperationCanceledException) when (!stoppingToken.IsCancellationRequested)
        {
            logger.LogError("Database retention cleanup exceeded its 30 minute time limit.");
        }
        catch (Exception exception) when (!stoppingToken.IsCancellationRequested)
        {
            logger.LogError(exception, "Database retention cleanup failed.");
        }
    }

    private static async Task<long> DeleteInBatchesAsync(
        NpgsqlConnection connection,
        string query,
        CancellationToken cancellationToken)
    {
        long total = 0;
        while (true)
        {
            await using var command = connection.CreateCommand();
            command.CommandText = query;
            command.Parameters.AddWithValue("batch_size", BatchSize);
            var deleted = await command.ExecuteNonQueryAsync(cancellationToken);
            total += deleted;
            if (deleted < BatchSize) return total;
            await Task.Delay(TimeSpan.FromMilliseconds(25), cancellationToken);
        }
    }
}
