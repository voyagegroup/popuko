import requests

from jewelpet.conf import settings

from .types import Repository, User, PullRequest, Issue, IssueComment, Branch

GITHUB_API = 'https://api.github.com'


def _request(method, path, data=None):
    """
    Request by any method

    Args:
        <string> HTTP method
        <string> request path
    Returns:
        <dict|None> response JSON
    """
    return getattr(requests, method)(
        '%s%s' % (GITHUB_API, path), json=data,
        headers={
            'Accept': 'application/vnd.github.v3+json',
            'Authorization': 'token %s' % settings['github']['token']
        },
        timeout=10)


def _put(path, data):
    """
    Request PUT method

    Args:
        <string> request path
        <dict> request parameters
    Returns:
        <dict|None> response JSON
    """
    res = requests.put(
        '%s%s' % (GITHUB_API, path), json=data,
        headers={
            'Accept': 'application/vnd.github.v3+json',
            'Authorization': 'token %s' % settings['github']['token']
        },
        timeout=10)
    return res


def _fill(params, keys):
    """
    Fill the dict keys

    Args:
        <dict> target dict
        <iterable> keys
    """
    for k in keys:
        if k not in params:
            params[k] = None


def get_repo(owner, repo_name):
    """
    Args:
        <string> owner
        <string> repository name
    Returns:
        <Repository|None>
    """
    res = _request('get', '/repos/%s/%s' % (owner, repo_name))
    if res.status_code != 200:
        return None
    params = res.json()
    _fill(params, ('parent', 'source', 'organization'))
    params['owner'] = _parse_user(params['owner'])
    return Repository(**params)


def _parse_user(params):
    """
    Args:
        <dict>
    Returns:
        <User>
    """
    for k in (
            'name',
            'company',
            'blog',
            'location',
            'email',
            'hireable',
            'bio',
            'public_repos',
            'public_gists',
            'followers',
            'following',
            'created_at',
            'updated_at'):
        if k not in params:
            params[k] = None
    return User(**params)


def get_user(name):
    """
    Args:
        <string> name
    Returns:
        <User|None>
    """
    res = _request('get', '/users/%s' % name)
    if res.status_code != 200:
        return None
    return User(**res.json())


def get_pr(owner, repo_name, pr_number):
    """
    Args:
        <string> owner
        <string> repository name
        <int> pr number
    Returns:
        <PullRequest|None>
    """
    res = _request('get', '/repos/%s/%s/pulls/%d' % (owner, repo_name, pr_number))
    if res.status_code != 200:
        return None
    params = res.json()
    params['links'] = params.pop('_links')  # namedtuple doesn't allow field name starts with a underscore
    _parse_issue(params)
    return PullRequest(**params)


def _parse_issue(params):
    """
    Args:
        <dict>
    """
    if params.get('user'):
        params['user'] = _parse_user(params['user'])
    if params.get('assignee'):
        params['assignee'] = _parse_user(params['assignee'])
    if params.get('assignees'):
        params['assignees'] = [_parse_user(x) for x in params['assignees']]


def get_issue(owner, repo_name, issue_number):
    """
    Args:
        <string> owner
        <string> repository name
        <int> issue number
    Returns:
        <Issue|None>
    """
    res = _request('get', '/repos/%s/%s/issues/%d' % (owner, repo_name, issue_number))
    if res.status_code != 200:
        return None
    params = res.json()
    _fill(params, ('pull_request',))
    _parse_issue(params)
    return Issue(**params)


def set_labels(owner, repo_name, issue_number, labels):
    """
    Args:
        <string> owner
        <string> repository name
        <int> issue number
        <iterable[string]> labels
    Returns:
        <bool> success or not
    """
    return edit_issue(labels=labels)


def assign(owner, repo_name, issue_number, assignees):
    """
    Args:
        <string> owner
        <string> repository name
        <int> issue number
        <iterable[string]> assignees
    Returns:
        <bool> success or not
    """
    return edit_issue(assignees=assignees)


def edit_issue(owner, repo_name, issue_number, **kwargs):
    res = _request('patch', '/repos/%s/%s/issues/%d' % (owner, repo_name, issue_number), kwargs)
    return res.status_code == 204


def get_branch(owner, repo_name, branch_name):
    """
    Args:
        <string> owner
        <string> repository name
        <string> branch name
    Returns:
        <Branch|None>
    """
    res = _request('get', '/repos/%s/%s/branches/%s' % (owner, repo_name, branch_name))
    if res.status_code != 200:
        return None
    params = res.json()
    params['links'] = params.pop('_links')  # namedtuple doesn't allow field name starts with a underscore
    return Branch(**params)


def create_branch(owner, repo_name, branch_name, base_sha):
    res = _request(
        'post',
        '/repos/%s/%s/git/refs' % (owner, repo_name),
        {'ref': 'refs/heads/%s' % branch_name, 'sha': base_sha})
    return res.status_code == 201


def delete_branch(owner, repo_name, branch_name):
    res = _request('delete', '/repos/%s/%s/git/refs/heads/%s' % (owner, repo_name, branch_name))
    return res.status_code == 204


def issue_comment(owner, repo_name, issue_number, comment):
    res = _request(
        'post',
        '/repos/%s/%s/issues/%d/comments' % (owner, repo_name, issue_number),
        {'body': comment})
    if res.status_code != 201:
        return None
    return IssueComment(**res)


def merge_pr(owner, repo_name, pr_number):
    _request('put', '/repos/%s/%s/pulls/%d/merge' % (owner, repo_name, pr_number), {})
