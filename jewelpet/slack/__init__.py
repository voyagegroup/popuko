import json

import requests

from jewelpet.conf import settings


def post(message):
    """
    Args:
        <string> message
    """
    params = dict(settings['slack']['params'])  # copy
    params['text'] = message
    return requests.post(
        settings['slack']['post_hook_url'],
        data=json.dumps(params),
        timeout=10)
