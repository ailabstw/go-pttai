#!/bin/bash

if [ "${BASH_ARGC}" != "1" ]
then
  virtualenv_dir="__"
else
  virtualenv_dir="${BASH_ARGV[0]}"
fi

the_basename=`basename \`pwd\``

echo "virtualenv_dir: ${virtualenv_dir} the_basename: ${the_basename}"

if [ ! -d ${virtualenv_dir} ]
then
  echo "no ${virtualenv_dir}. will create one"
  virtualenv -p `which python3` --prompt="[${the_basename}] " "${virtualenv_dir}"
fi

source ${virtualenv_dir}/bin/activate
the_python_path=`which python`
echo "python: ${the_python_path}"

echo "current_dir: "
pwd

# cookie-cutter
pip install -e git+https://github.com/chhsiao1981/cookiecutter.git@hsiao.skip-if-file-exists#egg=cookiecutter
