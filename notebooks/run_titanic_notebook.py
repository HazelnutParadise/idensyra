#!/usr/bin/env python3
"""
Run the Titanic notebook using nbconvert and save executed notebook.
"""
import subprocess, sys, os


def run(cmd):
    print('\n>>> Running:', cmd)
    ret = subprocess.run(cmd, shell=True, check=False, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)
    print(ret.stdout)
    if ret.returncode != 0:
        print('Command failed with exit code', ret.returncode)
        raise SystemExit(ret.returncode)


def pip_install(packages):
    run(f"{sys.executable} -m pip install --quiet {' '.join(packages)}")


if __name__ == "__main__":
    packages = ["nbconvert", "papermill", "seaborn", "pandas", "matplotlib", "scikit-learn"]
    pip_install(packages)
    input_nb = "notebooks/titanic_analysis.ipynb"
    output_nb = "notebooks/titanic_analysis_executed.ipynb"
    # Execute using nbconvert
    run(f"{sys.executable} -m nbconvert --to notebook --execute {input_nb} --output {output_nb} --ExecutePreprocessor.timeout=600")

    out_dir = "notebooks/outputs"
    print('\nOutput files in', out_dir)
    if os.path.exists(out_dir):
        for root, dirs, files in os.walk(out_dir):
            for f in files:
                print(os.path.join(root, f))
    else:
        print('No outputs directory found yet')

    print('\nDone.')
