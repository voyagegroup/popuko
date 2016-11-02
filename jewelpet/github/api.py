import requests

from jewelpet.conf import settings

from .types import User

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
