# El mismo contrato que el de Go: /suma hace CPU de verdad, /espera simula E/S. Un proceso de la BEAM por conexión.
defmodule Servidor.Plug do
  import Plug.Conn

  def init(opts), do: opts

  def call(%Plug.Conn{path_info: ["suma"]} = conn, _opts) do
    conn = fetch_query_params(conn)
    n = String.to_integer(conn.query_params["n"] || "0")
    t0 = System.monotonic_time(:nanosecond)
    s = suma(1, n, 0)
    ns = System.monotonic_time(:nanosecond) - t0
    json(conn, ~s({"lang":"elixir","n":#{n},"suma":#{s},"ns":#{ns}}))
  end

  def call(%Plug.Conn{path_info: ["espera"]} = conn, _opts) do
    conn = fetch_query_params(conn)
    ms = String.to_integer(conn.query_params["ms"] || "0")
    Process.sleep(ms)
    json(conn, ~s({"lang":"elixir","n":#{ms},"suma":0,"ns":0}))
  end

  def call(conn, _opts), do: send_resp(conn, 404, "")

  defp json(conn, cuerpo), do: conn |> put_resp_content_type("application/json") |> send_resp(200, cuerpo)

  defp suma(i, n, acc) when i > n, do: acc
  defp suma(i, n, acc), do: suma(i + 1, n, acc + rem(i * i, 1_000_003))
end

defmodule Servidor.Application do
  use Application

  def start(_type, _args) do
    {:ok, pid} =
      Supervisor.start_link([{Bandit, plug: Servidor.Plug, port: 8080, ip: {0, 0, 0, 0}}],
        strategy: :one_for_one)

    # La marca de «listo»: desde que arrancó la máquina virtual (BEAM) hasta escuchar.
    listo_us = System.convert_time_unit(System.monotonic_time() - :erlang.system_info(:start_time), :native, :microsecond)
    IO.puts(:stderr, "LISTO_EN_MS #{div(listo_us, 1000)}.#{String.pad_leading(Integer.to_string(rem(listo_us, 1000)), 3, "0")}")
    {:ok, pid}
  end
end
