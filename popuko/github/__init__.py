import hmac

from popuko.conf import settings

from popuko.github import api
from popuko.github.types import Repository, User, PullRequest, Issue, IssueComment, Branch


def is_valid_signature(body, api_signature):
    """
    Validate GitHub API signature

    Args:
        <string> request body
        <string> API request header 'X-Hub-Signature'
    Returns:
        <bool> request is valid or not
    """
    generated = hmac.new(
        bytes(settings['github']['hook_secret'], 'utf-8'),
        msg=body, digestmod='sha1').hexdigest()
    return 'sha1=%s' % generated == api_signature


def get_repo(owner, repo_name):
    """
    Args:
        <string> owner
        <string> repository name
    Returns:
        <Repository|None>
    """
    res = api._request('get', '/repos/%s/%s' % (owner, repo_name))
    if res.status_code != 200:
        return None
    params = res.json()
    api._fill(params, ('parent', 'source', 'organization'))
    params['owner'] = api._parse_user(params['owner'])
    return Repository(**params)


def get_user(name):
    """
    Args:
        <string> name
    Returns:
        <User|None>
    """
    res = api._request('get', '/users/%s' % name)
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
    res = api._request('get', '/repos/%s/%s/pulls/%d' % (owner, repo_name, pr_number))
    if res.status_code != 200:
        return None
    params = res.json()
    params['links'] = params.pop('_links')  # namedtuple doesn't allow field name starts with a underscore
    api._parse_issue(params)
    return PullRequest(**params)


def get_issue(owner, repo_name, issue_number):
    """
    Args:
        <string> owner
        <string> repository name
        <int> issue number
    Returns:
        <Issue|None>
    """
    res = api._request('get', '/repos/%s/%s/issues/%d' % (owner, repo_name, issue_number))
    if res.status_code != 200:
        return None
    params = res.json()
    api._fill(params, ('pull_request',))
    api._parse_issue(params)
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
    res = api._request('patch', '/repos/%s/%s/issues/%d' % (owner, repo_name, issue_number), kwargs)
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
    res = api._request('get', '/repos/%s/%s/branches/%s' % (owner, repo_name, branch_name))
    if res.status_code != 200:
        return None
    params = res.json()
    params['links'] = params.pop('_links')  # namedtuple doesn't allow field name starts with a underscore
    return Branch(**params)


def create_branch(owner, repo_name, branch_name, base_sha):
    res = api._request(
        'post',
        '/repos/%s/%s/git/refs' % (owner, repo_name),
        {'ref': 'refs/heads/%s' % branch_name, 'sha': base_sha})
    return res.status_code == 201


def delete_branch(owner, repo_name, branch_name):
    res = api._request('delete', '/repos/%s/%s/git/refs/heads/%s' % (owner, repo_name, branch_name))
    return res.status_code == 204


def issue_comment(owner, repo_name, issue_number, comment):
    res = api._request(
        'post',
        '/repos/%s/%s/issues/%d/comments' % (owner, repo_name, issue_number),
        {'body': comment})
    if res.status_code != 201:
        return None
    return IssueComment(**res)


def merge_pr(owner, repo_name, pr_number):
    api._request('put', '/repos/%s/%s/pulls/%d/merge' % (owner, repo_name, pr_number), {})
