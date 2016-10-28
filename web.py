from flask import Flask, request

import jewelpet.slack.commands
from jewelpet.conf import settings

app = Flask(__name__)
app.debug = True


@app.route('/slack', methods=['POST'])
def slack_app():
    args = request.form
    if args.get('token') != settings['slack']['token']:
        return '', 403

    (bot_name, command, *args) = args.get('text').split(' ')
    if bot_name != settings['bot_name']:
        return '{"text": "who?"}'
    method = getattr(jewelpet.slack.commands, command, None)
    if not method:
        return '{"text":"nothing"}'
    return '{"text":"%s"}' % method(*args)

if __name__ == '__main__':
    app.run(host='0.0.0.0')
