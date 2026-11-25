# Un contador como proceso supervisado. Lo hacemos fallar a propósito y miramos qué pasa.
defmodule Contador do
  use Agent
  def start_link(_), do: Agent.start_link(fn -> 0 end, name: __MODULE__)
end

{:ok, _} = Supervisor.start_link([Contador], strategy: :one_for_one)
IO.puts("contador vivo, pid #{inspect(Process.whereis(Contador))}")
Process.exit(Process.whereis(Contador), :kill)
Process.sleep(50)
IO.puts("tras matarlo, pid #{inspect(Process.whereis(Contador))}: el supervisor lo levantó")
IO.puts("el programa sigue: #{Agent.get(Contador, & &1)}")
