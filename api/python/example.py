from sona import api

try:
    api = api.SonaApi("lebron.xx.info")
except Exception as e:
    print(e)
    exit(1)

print(api.get("player", "team"))
print(api.get_list("friends", "list"))
