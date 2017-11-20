#!/usr/bin/env python
__copyright__ = '''
Copyright 2017 the original author or authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
'''
__author__ = 'David Turanski'

import importlib
import ntpath
import os
import os.path
import sys
import traceback
import zipfile
from shutil import copyfile
from urlparse import urlparse


def run_function(func):
    while True:
        try:
            data = raw_input()
            func(data)
        except KeyboardInterrupt:
            exit(0)
        except Exception:
            traceback.print_exc(file=sys.stdout)


def install_function():
    try:
        function_uri = os.environ['FUNCTION_URI']
        url = urlparse(function_uri)
        if (url.scheme == 'file'):
            if not os.path.isfile(url.path):
                sys.stderr.write("file %s does not exist\n" % url.path)
                exit(1)

            filename, extension = os.path.splitext(url.path)

            if (extension == '.zip'):
                zip_ref = zipfile.ZipFile(url.path, 'r')
                zip_ref.extractall('.')
                zip_ref.close()
                sys.stderr.write("Files extracted\n")
            elif extension == '.py':
                if not os.path.isfile(url.path):
                    copyfile(url.path, ('./%s' % ntpath.basename(url.path)))

        else:
            sys.stderr.write("scheme %s is not supported\n" % url.scheme)
            exit(1)

        indx = len('handler=')
        if len(url.query) <= indx or url.query[0:indx] != 'handler=':
            sys.stderr.write("FUNCTION_URI missing handler\n")
            exit(1)

        handler = url.query[indx:]
        if extension == '.py' and '.' not in handler:
            func_name = handler
            mod_name = ntpath.basename(filename)
        else:
            mod_name, func_name = handler.rsplit('.', 1)

        mod = importlib.import_module(mod_name)
        return getattr(mod, func_name)

    except KeyError:
        sys.stderr.write("required environment variable FUNCTION_URI is not defined\n")
        exit(1)


function = install_function()
run_function(function)
