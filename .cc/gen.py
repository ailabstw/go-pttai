#!/usr/bin/env python

import json
import sys
import logging

from cookiecutter.main import cookiecutter


def underscore_to_uppercase(the_str):
    return the_str.upper()


def underscore_to_camelcase(the_str):
    the_list = the_str.split('_')
    return ''.join([each_str.title() for each_str in the_list])

def underscore_to_lower_camelcase(the_str):
    the_list = the_str.split('_')
    return the_list[0] + ''.join([each_str.title() for each_str in the_list[1:]])

the_module = sys.argv[1]
full_name = sys.argv[2]

logging.warning('full_name: %s', full_name)

full_name_list = full_name.split('.')
if len(full_name_list) == 1:
    pkg = full_name_list[0]
    module = full_name_list[0]
    project = full_name_list[0]

    pkg_name = pkg
    project_name = project

    package_dir = '.'
else:
    pkg = full_name_list[-2]
    module = full_name_list[-1]
    project = full_name_list[-1]

    pkg_name = 'main' if full_name_list[0] == 'cmd' else pkg
    project_name = project

    package_dir = '/'.join(full_name_list[:-1])


the_dict = {
    'pkg': pkg,
    'module': module,
    # 'project': project,

    'pkg_name': pkg_name,
    'project_name': project_name,

    'package_dir': package_dir,

    'PKG': underscore_to_uppercase(pkg),
    'MODULE': underscore_to_uppercase(module),
    'PROJECT': underscore_to_camelcase(project),
    'PKG_NAME': underscore_to_uppercase(pkg_name),
    'PROJECT_NAME': underscore_to_camelcase(project_name),
    'PACKAGE_DIR': underscore_to_uppercase(package_dir),

    'Pkg': underscore_to_camelcase(pkg),
    'Module': underscore_to_camelcase(module),
    'Project': underscore_to_camelcase(project),
    'PkgName': underscore_to_camelcase(pkg_name),
    'ProjectName': underscore_to_camelcase(project_name),
    'PackageDir': underscore_to_camelcase(package_dir),

    'pkgLCamel': underscore_to_lower_camelcase(pkg),
    'moduleLCamel': underscore_to_lower_camelcase(module),
    'projectLCamel': underscore_to_lower_camelcase(project),
    'pkgName': underscore_to_lower_camelcase(pkg_name),
    'projectName': underscore_to_lower_camelcase(project_name),
    'packageDir': underscore_to_lower_camelcase(package_dir),
}

cookiecutter(
    '.cc/' + the_module,
    extra_context=the_dict,
    no_input=True,
    overwrite_if_exists=True,
    skip_if_file_exists=True,
)
