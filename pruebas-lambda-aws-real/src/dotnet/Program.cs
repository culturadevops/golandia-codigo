// Handler de prueba: EXACTAMENTE el mismo trabajo que en Go, Python y Node (vídeo 121).
// .NET con Native AOT: se compila a un binario `bootstrap` para el runtime provided.al2023.
using System.Text.Json.Serialization;
using Amazon.Lambda.Core;
using Amazon.Lambda.RuntimeSupport;
using Amazon.Lambda.Serialization.SystemTextJson;

public class Peticion
{
    [JsonPropertyName("n")] public int N { get; set; }
    [JsonPropertyName("io_ms")] public int IoMs { get; set; }
}

public class Respuesta
{
    [JsonPropertyName("lang")] public string Lang { get; set; } = "dotnet-aot";
    [JsonPropertyName("n")] public int N { get; set; }
    [JsonPropertyName("suma")] public ulong Suma { get; set; }
    [JsonPropertyName("io_ms")] public int IoMs { get; set; }
}

[JsonSerializable(typeof(Peticion))]
[JsonSerializable(typeof(Respuesta))]
public partial class Contexto : JsonSerializerContext { }

public static class Funcion
{
    public static async Task Main()
    {
        Func<Peticion, ILambdaContext, Task<Respuesta>> h = Manejar;
        await LambdaBootstrapBuilder.Create(h, new SourceGeneratorLambdaJsonSerializer<Contexto>())
            .Build()
            .RunAsync();
    }

    public static async Task<Respuesta> Manejar(Peticion p, ILambdaContext c)
    {
        ulong suma = 0;
        for (ulong i = 1; i <= (ulong)p.N; i++)
        {
            suma += i * i % 1000003;
        }
        if (p.IoMs > 0)
        {
            await Task.Delay(p.IoMs);
        }
        return new Respuesta { N = p.N, Suma = suma, IoMs = p.IoMs };
    }
}
