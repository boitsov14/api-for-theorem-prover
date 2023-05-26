import os

import requests


def notify_line(msg: str, path: str = None):
    print(msg)
    try:
        res = requests.post('https://notify-api.line.me/api/notify',
                            headers={'Authorization': 'Bearer ' + os.getenv('LINE_ACCESS_TOKEN')},
                            data={'message': msg if msg else 'Empty message.'},
                            files=({'imageFile': open(path, 'rb')} if path else None))
        if res.status_code != 200:
            print('LINE Notify Error:', res.text)
    except Exception as e:
        print('LINE Notify Error:', e)
