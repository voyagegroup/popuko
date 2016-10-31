import hmac

import github

from jewelpet.conf import settings
from jewelpet.exceptions import BranchConflictException


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


class Session(object):
    def __init__(self):
        """
        """
        self._g = github.Github(settings['github']['token'])
        self.user = self._g.get_user()

    def get_user(self, name):
        """
        Args:
            <string>
        Returns:
            <NamedUser>
        """
        return self._g.get_user(name)

    def get_organization(self, name):
        """
        Args:
            <string>
        Returns:
            <Organization>
        """
        return self._g.get_organization(name)

    def find_repo(self, req):
        """
        Args:
            <dict> request parameter of repository from GitHub API
                    {
                        "name": "repository name",
                        "owner": {
                            "login": "owner login",
                            "type": "User or Organization"
                        }
                    }
        Returns:
            <Repository>
        """
        owner_type = req['owner']['type']
        if owner_type not in ('User', 'Organization'):
            raise Exception('Unknown owner type "%s"' % owner_type)
        owner = getattr(self._g, 'get_%s' % owner_type.lower())(req['owner']['login'])
        return owner.get_repo(req['name'])


def build_auto(repo, pr_number, mode):
    """
    Args:
        <Repository>
        <number> number of PR
        <string> 'try' or 'r+'
    """
    if is_auto_branch_exists(repo):
        raise BranchConflictException('"auto" branch is already exists')

    pr = repo.get_issue(pr_number)
    pr.remove_from_labels('S-awaiting-review')
    pr.add_to_labels('S-awaiting-merge')

    head = repo.get_commit('HEAD')
    repo.create_git_ref('refs/heads/auto', head.sha)  # create auto branch

    repo.merge('auto', pr.head.sha, '%s %d' % (command, pr_number))
    # repo.merge('auto', pr.head.sha, '%s: auto merge branch "%s"' % (mode, pr.head.ref))


def is_auto_branch_exists(repo):
    """
    Returns:
        <bool>
    """
    try:
        if repo.get_branch('auto'):
            return True
    except github.GithubException as e:
        assert e.status == 404
        return False
