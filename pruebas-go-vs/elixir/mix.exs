# El servidor de la comparativa, en Elixir: Plug + Bandit (en la librería estándar de Elixir no hay un servidor HTTP
# de uso real). Se despliega como release de Mix, como en producción.
defmodule Servidor.MixProject do
  use Mix.Project

  def project do
    [app: :servidor, version: "1.0.0", elixir: "~> 1.20", start_permanent: true, deps: deps()]
  end

  def application do
    [extra_applications: [:logger], mod: {Servidor.Application, []}]
  end

  defp deps do
    [{:bandit, "1.13.1"}, {:plug, "1.20.3"}]
  end
end
