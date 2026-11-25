// El servidor de la comparativa, en C# (.NET 10): solo lo que trae la plataforma (ASP.NET Core, API mínima).
// El mismo contrato que el de Go: /suma hace CPU de verdad, /espera simula E/S. La misma fuente sirve para
// la versión normal (JIT) y para la compilada nativa (Native AOT).
using System.Diagnostics;

var arranque = Stopwatch.GetTimestamp();
var builder = WebApplication.CreateSlimBuilder(args);
builder.WebHost.UseUrls("http://0.0.0.0:8080");
builder.Logging.ClearProviders();
var app = builder.Build();

app.MapGet("/suma", (int n) =>
{
    var t0 = Stopwatch.GetTimestamp();
    ulong s = 0;
    for (int i = 1; i <= n; i++)
    {
        s += (ulong)i * (ulong)i % 1000003UL;
    }
    var ns = (long)(Stopwatch.GetElapsedTime(t0).TotalMilliseconds * 1_000_000);
    return Results.Text($"{{\"lang\":\"csharp\",\"n\":{n},\"suma\":{s},\"ns\":{ns}}}", "application/json");
});

app.MapGet("/espera", async (int ms) =>
{
    await Task.Delay(ms);
    return Results.Text($"{{\"lang\":\"csharp\",\"n\":{ms},\"suma\":0,\"ns\":0}}", "application/json");
});

// La marca de «listo» del propio proceso: main → escuchando, sin el contenedor por medio.
app.Lifetime.ApplicationStarted.Register(() =>
    Console.Error.WriteLine(FormattableString.Invariant($"LISTO_EN_MS {Stopwatch.GetElapsedTime(arranque).TotalMilliseconds:F3}")));
app.Run();
