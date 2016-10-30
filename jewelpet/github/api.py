import requests

from jewelpet.conf import settings

from .types import Repository, User

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


def get_repo(owner, repo_name):
    """
    Args:
        <string> owner
        <string> repository name
    """
    res = _get('/repos/%s/%s' % (owner, repo_name))
    for k in ('parent', 'source', 'organization'):
        if k not in res:
            res[k] = None
    res['owner'] = _parse_user(res['owner'])
    repo = Repository(**res)
    return repo


def get_user(name):
    """
    Args:
        <string> name
    """
    res = _get('/users/%s' % name)
    user = User(**res)
    return user


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
    user = User(**params)
    return user


def get_pr(owner, repo_name, pr_number):
    """
    Args:
        <string> owner
        <string> repository name
        <int> pr number
    """
    res = _get('/users/%s/%s/pulls/%d' % (owner, repo_name, pr_number))
    user = User(**res)
    return user
