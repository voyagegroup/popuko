import requests

from jewelpet.conf import settings

from .types import Repository, User, PullRequest, Issue

GITHUB_API = 'https://api.github.com'


def _get(path):
    """
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


def _fill(params, keys):
    """
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
        <Repository>
    """
    res = _get('/repos/%s/%s' % (owner, repo_name))
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
        <User>
    """
    res = _get('/users/%s' % name)
    return User(**res)


def get_pr(owner, repo_name, pr_number):
    """
    Args:
        <string> owner
        <string> repository name
        <int> pr number
    """
    res = _get('/repos/%s/%s/pulls/%d' % (owner, repo_name, pr_number))
    res['links'] = res.pop('_links')  # namedtuple doesn't allow field name starts with a underscore
    return PullRequest(**res)


def get_issue(owner, repo_name, issue_number):
    """
    Args:
        <string> owner
        <string> repository name
        <int> issue number
    """
    res = _get('/repos/%s/%s/issues/%d' % (owner, repo_name, issue_number))
    _fill(res, ('pull_request',))
    return Issue(**res)
