# import websocket
# import threading
# import json
# import random

# def on_message(ws, message):
#     msg = json.loads(message)
#     gameData = msg['gameData']
#     if msg['error'] != '':
#         print(msg['error'])
#         return
    
#     print(msg['action'])
#     if gameData['status'] == 'waiting':
#         print(gameData)
#     if gameData['status'] == 'playing':
#         print("current stack:", gameData['stackValue'])
#         print("current player:", gameData['currentPlayerIndex'])
#         for card in gameData['playerCards']:
#             print(card)

# def on_error(ws, error):
#     print('????')
#     print(f"Error: {error}")

# def on_close(ws, _, __):
#     print("Connection closed")

# def on_open(ws):
#     print("Connection opened")
#     #random user Id
#     userId = random.randint(0, 100000000)
#     # Send a message after connection is opened
#     action = {
#         "action": "join",
#         "userId": str(userId),
#         "username": "test" + str(random.randint(0, 1000)),
#         "profilePic": ""
#     }
#     ws.send(json.dumps(action))

# global ws

# def start_websocket_client():
#     global ws
#     ws = websocket.WebSocketApp("ws://localhost:8080/ws/000000000000",
#                                 on_message=on_message,
#                                 on_error=on_error,
#                                 on_close=on_close)
#     ws.on_open = on_open

#     ws.run_forever()

# if __name__ == "__main__":
#     # Start WebSocket client in a separate thread
#     client_thread = threading.Thread(target=start_websocket_client)
#     client_thread.start()

#     while True:
#         msg = input()
#         if msg == 'q':
#             ws.close()
#             break
        
#         if msg == '1':
#             action = {
#                 "action": "start"
#             }
#             print('start')
#             ws.send(json.dumps(action))

#         if msg == '2':
#             value = input('value: ')
#             isSpecial = input('isSpecial: ') == '1'
#             action = {
#                 "action": "play",
#                 "card": {
#                     "value": int(value),
#                     "isSpecial": isSpecial
#                 }
#             }
#             print('play')
#             ws.send(json.dumps(action))

import requests

url = "http://localhost:8080"

# edit this
body_data = {
    "userId": "whuf80z9sGMkkSNSuDYydvcIA3q1",
}

# edit this
path = "/createroom"

if __name__ == '__main__':
    res = requests.post(f"{url}{path}", json=body_data)
    print(res.status_code)
    print(res.content)