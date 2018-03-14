import collections

def concat(vals):
    '''
    :param vals: expects a dict
    :return: a singleton dict whose value is concatenated keys and values
    '''
    od = collections.OrderedDict(sorted(vals.items()))
    result = ''
    for k, v in od.items():
        result = result + k + v
    return {'result': result}
