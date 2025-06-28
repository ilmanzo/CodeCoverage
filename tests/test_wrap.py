import os
import tempfile
import shutil
import stat
import unittest
from unittest import mock

import sys
sys.path.append("..")
import wrap

class TestWrapScript(unittest.TestCase):
    def setUp(self):
        # Create a temp directory for SAFE_BIN_DIR and LOG_DIR
        self.test_dir = tempfile.mkdtemp()
        self.safe_bin_dir = os.path.join(self.test_dir, "bin")
        self.log_dir = os.path.join(self.test_dir, "data")
        self.pin_tool_dir = os.path.join(self.test_dir, "pin_tool")
        os.makedirs(self.safe_bin_dir)
        os.makedirs(self.log_dir)
        os.makedirs(self.pin_tool_dir)
        # Create a dummy FuncTracer.so
        self.func_tracer_path = os.path.join(self.pin_tool_dir, "FuncTracer.so")
        with open(self.func_tracer_path, "w") as f:
            f.write("dummy")
        # Patch configuration in wrap.py
        wrap.SAFE_BIN_DIR = self.safe_bin_dir
        wrap.LOG_DIR = self.log_dir
        wrap.PIN_TOOL_SEARCH_DIR = self.test_dir
        wrap.PIN_ROOT = self.test_dir

        # Create a dummy binary file
        self.binary_path = os.path.join(self.test_dir, "dummy_binary")
        with open(self.binary_path, "w") as f:
            f.write("#!/bin/bash\necho Hello\n")
        os.chmod(self.binary_path, 0o755)

    def tearDown(self):
        shutil.rmtree(self.test_dir)

    def test_find_pin_tool(self):
        found = wrap.find_pin_tool()
        self.assertEqual(found, self.func_tracer_path)

    def test_install_and_uninstall_regular_file(self):
        wrap.handle_install(self.binary_path, self.func_tracer_path)
        # After install, the binary should be a wrapper script
        with open(self.binary_path) as f:
            content = f.read()
        self.assertIn(wrap.WRAPPER_ID_COMMENT, content)
        # Uninstall should restore the original binary
        wrap.handle_uninstall(self.binary_path)
        with open(self.binary_path) as f:
            content = f.read()
        self.assertIn("echo Hello", content)

    def test_install_symlink(self):
        # Create a symlink to the binary
        symlink_path = os.path.join(self.test_dir, "symlink_binary")
        os.symlink(self.binary_path, symlink_path)
        wrap.handle_install(symlink_path, self.func_tracer_path)
        # The original binary should be replaced, not the symlink
        with open(self.binary_path) as f:
            content = f.read()
        self.assertIn(wrap.WRAPPER_ID_COMMENT, content)
        # The symlink should still exist and point to the same file
        self.assertTrue(os.path.islink(symlink_path))
        self.assertEqual(os.readlink(symlink_path), self.binary_path)

    def test_install_already_wrapped(self):
        wrap.handle_install(self.binary_path, self.func_tracer_path)
        with self.assertRaises(SystemExit):
            wrap.handle_install(self.binary_path, self.func_tracer_path)

    def test_uninstall_not_wrapper(self):
        with self.assertRaises(SystemExit):
            wrap.handle_uninstall(self.binary_path)

if __name__ == "__main__":