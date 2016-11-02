import hmac

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
