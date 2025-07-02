import unittest
import sys
sys.path.append("..")  # Adjust path to import the module
from coverage_analyzer import analyze_logs

class TestCoverageAnalyzer(unittest.TestCase):
    def setUp(self):
        self.log_filename = "mock_log.txt"

    def tearDown(self):
        import os
        if os.path.exists(self.log_filename):
            os.remove(self.log_filename)

    def write_log(self, lines):
        with open(self.log_filename, "w") as f:
            for line in lines:
                f.write(line + "\n")

    def test_basic_coverage(self):
        self.write_log([
            "[Image:myprog] [Function:foo]",
            "[Image:myprog] [Function:bar]",
            "[Image:myprog] [Called:foo]",
            "[Image:myprog] [Function:baz]",
        ])
        data = analyze_logs([self.log_filename])
        self.assertIn("myprog", data)
        self.assertEqual(data["myprog"]["total_functions"], {"foo", "bar", "baz"})
        self.assertEqual(data["myprog"]["called_functions"], {"foo"})

    def test_no_functions(self):
        self.write_log([
            "[Image:myprog] [Section:.text]",
        ])
        data = analyze_logs([self.log_filename])
        self.assertEqual(data["myprog"]["total_functions"], set())
        self.assertEqual(data["myprog"]["called_functions"], set())

    def test_no_called_functions(self):
        self.write_log([
            "[Image:myprog] [Function:foo]",
            "[Image:myprog] [Function:bar]",
        ])
        data = analyze_logs([self.log_filename])
        self.assertEqual(data["myprog"]["total_functions"], {"foo", "bar"})
        self.assertEqual(data["myprog"]["called_functions"], set())

    def test_called_function_not_in_total(self):
        self.write_log([
            "[Image:myprog] [Called:foo]",
        ])
        data = analyze_logs([self.log_filename])
        self.assertEqual(data["myprog"]["total_functions"], set())
        self.assertEqual(data["myprog"]["called_functions"], {"foo"})

    def test_multiple_images(self):
        self.write_log([
            "[Image:prog1] [Function:foo]",
            "[Image:prog2] [Function:bar]",
            "[Image:prog1] [Called:foo]",
            "[Image:prog2] [Called:bar]",
        ])
        data = analyze_logs([self.log_filename])
        self.assertIn("prog1", data)
        self.assertIn("prog2", data)
        self.assertEqual(data["prog1"]["total_functions"], {"foo"})
        self.assertEqual(data["prog1"]["called_functions"], {"foo"})
        self.assertEqual(data["prog2"]["total_functions"], {"bar"})
        self.assertEqual(data["prog2"]["called_functions"], {"bar"})

    def test_duplicate_entries(self):
        self.write_log([
            "[Image:myprog] [Function:foo]",
            "[Image:myprog] [Function:foo]",
            "[Image:myprog] [Called:foo]",
            "[Image:myprog] [Called:foo]",
        ])
        data = analyze_logs([self.log_filename])
        self.assertEqual(data["myprog"]["total_functions"], {"foo"})
        self.assertEqual(data["myprog"]["called_functions"], {"foo"})

    def test_irrelevant_lines(self):
        self.write_log([
            "Random log line",
            "[Image:myprog] [Section:.text]",
            "[Image:myprog] [Function:foo]",
            "[Image:myprog] [Called:foo]",
            "Another random line"
        ])
        data = analyze_logs([self.log_filename])
        self.assertEqual(data["myprog"]["total_functions"], {"foo"})
        self.assertEqual(data["myprog"]["called_functions"], {"foo"})

if __name__ == "__main__":
    unittest.main()