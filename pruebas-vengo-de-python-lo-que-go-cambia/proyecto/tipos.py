# Un tipo equivocado: el puerto llega como texto desde un archivo de configuración.
def siguiente_puerto(puerto):
    return puerto + 1

config = {"puerto": "8080"}          # leído de un archivo: es texto
print("arranca")
print(siguiente_puerto(config["puerto"]))
