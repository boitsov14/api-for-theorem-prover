import base64
import logging
import os
import random
import string
from tempfile import TemporaryDirectory

from dotenv import load_dotenv
from flask import Flask, request
from flask_cors import CORS
from misskey import Misskey, NoteVisibility
from waitress import serve

from notify import notify_line
from prover import make_proof_tree

app = Flask(__name__)
CORS(app)
logging.getLogger('waitress').setLevel(logging.INFO)


@app.route('/web', methods=['POST'])
def web_app():
    txt = request.json['txt']
    notify_line('Web: ' + txt)
    with TemporaryDirectory(dir=os.getcwd()) as work:
        try:
            os.chdir(work)
            os.symlink('../prover', 'prover')
            msg = make_proof_tree(txt, '500m', 10)
            res = {'msg': msg}
            if os.path.exists('out.png'):
                notify_line(msg, 'out.png')
                with open('out.png', 'rb') as f:
                    res['img'] = base64.b64encode(f.read()).decode()
                with open('out.tex', 'r') as f:
                    res['tex'] = f.read()
            else:
                notify_line(msg)
        except Exception as e:
            notify_line(f'Unexpected error has occurred: {e}')
            res = {'msg': 'Unexpected error has occurred.'}
        finally:
            os.chdir('..')
    return res


@app.route('/misskey', methods=['POST'])
def misskey_app():
    # 認証確認
    if request.headers.get('Authorization') != 'Bearer ' + os.getenv('PASSWORD'):
        notify_line('Unauthorized request has been detected.')
        return 'Unauthorized', 401
    note_id, username, txt = request.json['id'], request.json['username'], request.json['txt']
    notify_line('Misskey: ' + txt)
    with TemporaryDirectory(dir=os.getcwd()) as work:
        try:
            os.chdir(work)
            os.symlink('../prover', 'prover')
            msg = make_proof_tree(txt, '2g', 30)
            api = Misskey(i=os.getenv('MISSKEY_ACCESS_TOKEN'))
            if os.path.exists('out.png'):
                notify_line(msg, 'out.png')
                # 画像アップロード
                with open('out.png', 'rb') as f:
                    file_id = api.drive_files_create(f)['id']
                # Note投稿
                res = api.notes_create(text=username + ' ' + msg, renote_id=note_id, file_ids=[file_id],
                                       visibility=NoteVisibility.HOME)
                created_id = res['createdNote']['id']
                notify_line(f'https://misskey.io/notes/{created_id}')
            else:
                notify_line(msg)
                if 'seconds' not in msg:
                    msg += ' [' + ''.join(random.sample(string.ascii_lowercase, 3)) + ']'
                # Note投稿
                res = api.notes_create(text=username + ' ' + msg, renote_id=note_id, visibility=NoteVisibility.HOME)
                created_id = res['createdNote']['id']
                notify_line(f'https://misskey.io/notes/{created_id}')
        except Exception as e:
            notify_line(f'Unexpected error has occurred: {e}')
        finally:
            os.chdir('..')
    return 'OK', 200


if __name__ == '__main__':
    load_dotenv()
    serve(app, port=os.getenv('PORT', 3000), threads=1)
