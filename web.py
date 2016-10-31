# -*- coding: utf-8 -*-

import json
import re

from flask import Flask, request

import jewelpet.slack.commands
from jewelpet import github, slack
from jewelpet.conf import settings
from jewelpet.exceptions import BranchConflictException

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

        s = github.Session()
        repo = s.find_repo(req['repository'])

        (trigger, command, *args) = req['comment']['body'].split(' ')
        if trigger in settings['github']['reviewers'] and command == 'r?':
            issue = repo.get_issue(int(req['issue']['number']))
            issue.edit(assignee=s.get_user(trigger[1:]))
            issue.add_to_labels('S-awaiting-review')
            return 'ASSIGNED'

        if trigger != settings['github']['trigger']:
            return 'PASS'

        if command in ('try', 'r+'):
            pr_number = int(req['issue']['number'])
            github.build_auto(repo, pr_number, command)
            slack.post('created "auto" branch')
            return 'AUTO'
    except BranchConflictException:
        slack.post('"auto" branch is already exists')
        return 'AUTO ALREADY EXISTS'
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

    s = github.Session()
    owner = s.get_organization(req['repository']['owner_name'])
    repo = owner.get_repo(req['repository']['name'])

    # state = (passed|failed)
    state = req.get('state', '')
    if state == 'passed':
        slack.post('"auto"のビルドが成功したようだな')
        ref = repo.get_git_ref('heads/auto')
        ref.delete()
        slack.post('autoブランチ消したった')
        m = re.match(r'^r\+ (\d+)', req.get('message', ''))
        if m:
            slack.post('mergeするぞ')
            pr_number = int(m.groups()[0])
            pr = repo.get_pull(pr_number)
            pr.merge()
            slack.post('mergeした')
            repo.get_git_ref('heads/%s' % pr.head.ref).delete()
            slack.post('%sブランチ消した')
    elif state == 'failed':
        slack.post('"auto"のビルドは失敗したようだな')
        ref = repo.get_git_ref('heads/auto')
        ref.delete()
        slack.post('autoブランチ消したった')
    else:
        slack.post('"auto"のビルドがよく分からん')


if __name__ == '__main__':
    app.run(host='0.0.0.0')
