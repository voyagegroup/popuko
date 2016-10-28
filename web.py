# -*- coding: utf-8 -*-

import json

from flask import Flask, request

import jewelpet.slack.commands
from jewelpet import github, slack
from jewelpet.conf import settings

app = Flask(__name__)
app.debug = True


@app.route('/slack', methods=['POST'])
def slack_app():
    args = request.form
    if args.get('token') != settings['slack']['token']:
        return '', 403

    (bot_name, command, *args) = args.get('text').split(' ')
    if bot_name != settings['slack']['bot_name']:
        return '{"text": "who?"}'
    method = getattr(jewelpet.slack.commands, command, None)
    if not method:
        return '{"text":"nothing"}'
    return '{"text":"%s"}' % method(*args)


@app.route('/github', methods=['POST'])
def github_app():
    req = request.json
    if not github.is_valid_signature(
            request.data,
            request.headers.get('X-Hub-Signature')):
        return 'Invalid Signature'
    try:
        if req['action'] != 'created':
            return 'PASS'
        (trigger, command, *args) = req['comment']['body'].split(' ')

        if trigger in settings['github']['reviewers'] and command == 'r?':
            s = github.Session()
            repo = s.find_repo(req['repository'])
            issue = repo.get_issue(int(req['issue']['number']))
            issue.edit(assignee=s.get_user(trigger[1:]))
            return 'ASSIGNED'

        if trigger != settings['github']['trigger']:
            return 'PASS'
        if command == 'try':
            s = github.Session()
            repo = s.find_repo(req['repository'])
            if github.is_auto_branch_exists(repo):
                slack.post('"auto" branch is already exists')
                return 'AUTO ALREADY EXISTS'

            pr_number = int(req['issue']['number'])
            github.build_auto(repo, pr_number, 'try')
            slack.post('created "auto" branch')
            return 'AUTO'
    except ValueError:
        return 'ValueError'
    except KeyError:
        return 'KeyError'


@app.route('/travis', methods=['POST'])
def travis_app():
    req = json.loads(request.form.get('payload'))
    from pprint import pprint
    pprint(req)
    if req.get('branch', '') != 'auto':
        return 'PASS'

    # state = (passed|failed)
    state = req.get('state', '')
    if state == 'passed':
        slack.post('"auto"のビルドが成功したようだな')
        return 'OK'
    elif state == 'failed':
        slack.post('"auto"のビルドは失敗したようだな')
        return 'NG'
    else:
        slack.post('"auto"のビルドがよく分からん')

if __name__ == '__main__':
    app.run(host='0.0.0.0')
