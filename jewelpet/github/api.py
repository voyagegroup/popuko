import requests

from jewelpet.conf import settings

from .types import Repository, User, PullRequest, Issue, Branch

GITHUB_API = 'https://api.github.com'


def _get(path):
    """
    Request GET method

    Args:
        <string> request path
    Returns:
        <dict|None> response JSON
    """
    res = requests.get(
        '%s%s' % (GITHUB_API, path),
        headers={
            'Accept': 'application/vnd.github.v3+json',
            'Authorization': 'token %s' % settings['github']['token']
        },
        timeout=10)
    if res.status_code != 200:
        return None
    return res.json()


def _post(path, data):
    """
    Request POST method

    Args:
        <string> request path
        <dict> request parameters
    Returns:
        <dict|None> response JSON
    """
    res = requests.post(
        '%s%s' % (GITHUB_API, path), json=data,
        headers={
            'Accept': 'application/vnd.github.v3+json',
            'Authorization': 'token %s' % settings['github']['token']
        },
        timeout=10)
    if res.status_code != 201:
        return None
    return res.json()


def _patch(path, data):
    """
    Request PATCH method

    Args:
        <string> request path
        <dict> request parameters
    Returns:
        <dict|None> response JSON
    """
    res = requests.patch(
        '%s%s' % (GITHUB_API, path), json=data,
        headers={
            'Accept': 'application/vnd.github.v3+json',
            'Authorization': 'token %s' % settings['github']['token']
        },
        timeout=10)
    if res.status_code != 200:
        return None
    return res.json()


def _delete(path):
    """
    Request DELETE method

    Args:
        <string> request path
    Returns:
        <bool> success or not
    """
    res = requests.delete(
        '%s%s' % (GITHUB_API, path),
        headers={
            'Accept': 'application/vnd.github.v3+json',
            'Authorization': 'token %s' % settings['github']['token']
        },
        timeout=10)
    return res.status_code == 204


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
    res = _get('/repos/%s/%s' % (owner, repo_name))
    if res is None:
        return None
    _fill(res, ('parent', 'source', 'organization'))
    res['owner'] = _parse_user(res['owner'])
    return Repository(**res)


def _parse_user(params):
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
    res = _get('/users/%s' % name)
    if res is None:
        return None
    return User(**res)


def get_pr(owner, repo_name, pr_number):
    """
    Args:
        <string> owner
        <string> repository name
        <int> pr number
    Returns:
        <PullRequest|None>
    """
    res = _get('/repos/%s/%s/pulls/%d' % (owner, repo_name, pr_number))
    if res is None:
        return None
    res['links'] = res.pop('_links')  # namedtuple doesn't allow field name starts with a underscore
    pr = PullRequest(**res)
    pr_ = pr._replace(
        user=_parse_user(pr.user),
        assignee=_parse_user(pr.assignee),
        assignees=[_parse_user(x) for x in pr.assignees])
    return pr_


def get_issue(owner, repo_name, issue_number):
    """
    Args:
        <string> owner
        <string> repository name
        <int> issue number
    Returns:
        <Issue|None>
    """
    res = _get('/repos/%s/%s/issues/%d' % (owner, repo_name, issue_number))
    if res is None:
        return None
    _fill(res, ('pull_request',))
    issue = Issue(**res)
    issue_ = issue._replace(
        user=_parse_user(issue.user),
        assignee=_parse_user(issue.assignee),
        assignees=[_parse_user(x) for x in issue.assignees])
    return issue_


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
    res = _patch('/repos/%s/%s/issues/%d' % (owner, repo_name, issue_number), {'labels': labels})
    return bool(res)


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
    res = _patch('/repos/%s/%s/issues/%d' % (owner, repo_name, issue_number), {'assignees': assignees})
    return bool(res)


def get_branch(owner, repo_name, branch_name):
    """
    Args:
        <string> owner
        <string> repository name
        <string> branch name
    Returns:
        <Branch|None>
    """
    res = _get('/repos/%s/%s/branches/%s' % (owner, repo_name, branch_name))
    if res is None:
        return None
    res['links'] = res.pop('_links')  # namedtuple doesn't allow field name starts with a underscore
    return Branch(**res)


def create_branch(owner, repo_name, branch_name, base_sha):
    return _post('/repos/%s/%s/git/refs' % (owner, repo_name), {'ref': 'refs/heads/%s' % branch_name, 'sha': base_sha})


def delete_branch(owner, repo_name, branch_name):
    return _delete('/repos/%s/%s/git/refs/heads/%s' % (owner, repo_name, branch_name))
