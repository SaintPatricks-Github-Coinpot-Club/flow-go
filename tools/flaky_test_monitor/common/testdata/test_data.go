package testdata

import (
	"time"

	"github.com/onflow/flow-go/tools/flaky_test_monitor/common"
)

// TestDataLevel1 is used by tests to store what the expected test result should be and what the raw
// JSON input file is
type TestDataLevel1 struct {
	ExpectedTestRun    common.TestRun
	RawJSONTestRunFile string
}

type TestDataLevel2 struct {
	Directory        string
	HasFailures      bool
	HasNoResultTests bool
	TestRuns         []common.TestRun
}

// ************** Helper Functions *****************
// following functions are used to construct expected TestRun data

func getCommitDate() time.Time {
	commitDate, err := time.Parse(time.RFC3339, "2021-09-22T01:06:25Z")
	common.AssertNoError(err, "time parse - commit date")
	return commitDate
}

func getJobRunDate() time.Time {
	jobRunDate, err := time.Parse(time.RFC3339, "2021-09-22T04:06:25Z")
	common.AssertNoError(err, "time parse - job run date")
	return jobRunDate
}

func getCommitSha() string {
	return "46baf6c6be29af9c040bc14195e195848598bbae"
}

func getCryptoHashPackage() string {
	return "github.com/onflow/flow-go/crypto/hash"
}

func getPassedTest(name string) common.TestResultRow {
	return getPassedTestPackage(name, getCryptoHashPackage())
}

func getPassedTestPackage(name string, packageName string) common.TestResultRow {
	row := common.TestResultRow{
		TestResult: common.TestResult{
			CommitSha:  getCommitSha(),
			CommitDate: getCommitDate(),
			JobRunDate: getJobRunDate(),
			Test:       name,
			Package:    packageName,
			Result:     "1",
			Elapsed:    0,
			Output: []struct {
				Item string "json:\"item\""
			}{
				{Item: "=== RUN   " + name + "\n"},
				{Item: "--- PASS: " + name + " (0.00s)\n"},
			},
		},
	}
	return row
}

func getPassedTestElapsed(name string, elapsed float32, elapsedStr string) common.TestResultRow {
	row := common.TestResultRow{
		TestResult: common.TestResult{
			CommitSha:  getCommitSha(),
			CommitDate: getCommitDate(),
			JobRunDate: getJobRunDate(),
			Test:       name,
			Package:    getCryptoHashPackage(),
			Result:     "1",
			Elapsed:    elapsed,
			Output: []struct {
				Item string "json:\"item\""
			}{
				{Item: "=== RUN   " + name + "\n"},
				{Item: "    --- PASS: " + name + " (" + elapsedStr + "s)\n"},
			},
		},
	}
	return row
}

func getPassedTestPackageElapsedOutput(name string, packageName string, elapsed float32, elapsedStr string, output string) common.TestResultRow {
	row := getPassedTestElapsedOutput(name, elapsed, elapsedStr, output)
	row.TestResult.Package = packageName
	return row
}

func getPassedTestElapsedOutput(name string, elapsed float32, elapsedStr string, output string) common.TestResultRow {
	row := common.TestResultRow{
		TestResult: common.TestResult{
			CommitSha:  getCommitSha(),
			CommitDate: getCommitDate(),
			JobRunDate: getJobRunDate(),
			Test:       name,
			Package:    getCryptoHashPackage(),
			Result:     "1",
			Elapsed:    elapsed,
			Output: []struct {
				Item string "json:\"item\""
			}{
				{Item: "=== RUN   " + name + "\n"},
				{Item: output},
				{Item: "--- PASS: " + name + " (" + elapsedStr + "s)\n"},
			},
		},
	}
	return row
}

func getFailedTest_TestSanitySha2_256() common.TestResultRow {
	row := common.TestResultRow{
		TestResult: common.TestResult{
			CommitSha:  getCommitSha(),
			CommitDate: getCommitDate(),
			JobRunDate: getJobRunDate(),
			Test:       "TestSanitySha2_256",
			Package:    getCryptoHashPackage(),
			Result:     "0",
			Elapsed:    0,
			Output: []struct {
				Item string "json:\"item\""
			}{
				{Item: "=== RUN   TestSanitySha2_256\n"},
				{Item: "    hash_test.go:41: \n"},
				{Item: "        \tError Trace:\thash_test.go:41\n"},
				{Item: "        \tError:      \tNot equal: \n"},
				{Item: "        \t            \texpected: hash.Hash{0x9f, 0x86, 0xd0, 0x81, 0x88, 0x4c, 0x7d, 0x65, 0x9a, 0x2f, 0xea, 0xa0, 0xc5, 0x5a, 0xd0, 0x15, 0xa3, 0xbf, 0x4f, 0x1b, 0x2b, 0xb, 0x82, 0x2c, 0xd1, 0x5d, 0x6c, 0x15, 0xb0, 0xf0, 0xa, 0x9}\n"},
				{Item: "        \t            \tactual  : hash.Hash{0x9f, 0x86, 0xd0, 0x81, 0x88, 0x4c, 0x7d, 0x65, 0x9a, 0x2f, 0xea, 0xa0, 0xc5, 0x5a, 0xd0, 0x15, 0xa3, 0xbf, 0x4f, 0x1b, 0x2b, 0xb, 0x82, 0x2c, 0xd1, 0x5d, 0x6c, 0x15, 0xb0, 0xf0, 0xa, 0x8}\n"},
				{Item: "        \t            \t\n"},
				{Item: "        \t            \tDiff:\n"},
				{Item: "        \t            \t--- Expected\n"},
				{Item: "        \t            \t+++ Actual\n"},
				{Item: "        \t            \t@@ -1,2 +1,2 @@\n"},
				{Item: "        \t            \t-(hash.Hash) (len=32) 0x9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a09\n"},
				{Item: "        \t            \t+(hash.Hash) (len=32) 0x9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08\n"},
				{Item: "        \t            \t \n"},
				{Item: "        \tTest:       \tTestSanitySha2_256\n"},
				{Item: "--- FAIL: TestSanitySha2_256 (0.00s)\n"},
			},
		},
	}
	return row
}

func getFailedTest_TestSanitySha3_256() common.TestResultRow {
	row := common.TestResultRow{
		TestResult: common.TestResult{
			CommitSha:  getCommitSha(),
			CommitDate: getCommitDate(),
			JobRunDate: getJobRunDate(),
			Test:       "TestSanitySha3_256",
			Package:    getCryptoHashPackage(),
			Result:     "0",
			Elapsed:    0,
			Output: []struct {
				Item string "json:\"item\""
			}{
				{Item: "=== RUN   TestSanitySha3_256\n"},
				{Item: "    hash_test.go:21: \n"},
				{Item: "        \tError Trace:\thash_test.go:21\n"},
				{Item: "        \tError:      \tNot equal: \n"},
				{Item: "        \t            \texpected: hash.Hash{0x36, 0xf0, 0x28, 0x58, 0xb, 0xb0, 0x2c, 0xc8, 0x27, 0x2a, 0x9a, 0x2, 0xf, 0x42, 0x0, 0xe3, 0x46, 0xe2, 0x76, 0xae, 0x66, 0x4e, 0x45, 0xee, 0x80, 0x74, 0x55, 0x74, 0xe2, 0xf5, 0xab, 0x81}\n"},
				{Item: "        \t            \tactual  : hash.Hash{0x36, 0xf0, 0x28, 0x58, 0xb, 0xb0, 0x2c, 0xc8, 0x27, 0x2a, 0x9a, 0x2, 0xf, 0x42, 0x0, 0xe3, 0x46, 0xe2, 0x76, 0xae, 0x66, 0x4e, 0x45, 0xee, 0x80, 0x74, 0x55, 0x74, 0xe2, 0xf5, 0xab, 0x80}\n"},
				{Item: "        \t            \t\n"},
				{Item: "        \t            \tDiff:\n"},
				{Item: "        \t            \t--- Expected\n"},
				{Item: "        \t            \t+++ Actual\n"},
				{Item: "        \t            \t@@ -1,2 +1,2 @@\n"},
				{Item: "        \t            \t-(hash.Hash) (len=32) 0x36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab81\n"},
				{Item: "        \t            \t+(hash.Hash) (len=32) 0x36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab80\n"},
				{Item: "        \t            \t \n"},
				{Item: "        \tTest:       \tTestSanitySha3_256\n"},
				{Item: "--- FAIL: TestSanitySha3_256 (0.00s)\n"},
			},
		},
	}
	return row
}

func getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack() common.TestResultRow {
	row := common.TestResultRow{
		TestResult: common.TestResult{
			CommitSha:  getCommitSha(),
			CommitDate: getCommitDate(),
			JobRunDate: getJobRunDate(),
			Test:       "TestEncodableRandomBeaconPrivKeyMsgPack",
			Package:    "github.com/onflow/flow-go/model/encodable",
			Result:     "-100",
			Elapsed:    0,
			Output: []struct {
				Item string "json:\"item\""
			}{
				{Item: "=== RUN   TestEncodableRandomBeaconPrivKeyMsgPack\n"},
				{Item: "bytes: 194--- PASS: TestEncodableRandomBeaconPrivKeyMsgPack (0.00s)\n"},
			},
		},
	}
	return row
}

// ************** Level 1 - Expected Test Run Functions *****************
// following functions are used by unit tests for constructing expected TestRun data

func GetTestData_Level1_1CountSingleNilTest() common.TestRun {
	row1 := getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack()

	testRun := common.TestRun{
		Rows: []common.TestResultRow{
			row1,
		},
	}
	return testRun
}

func GetTestData_Level1_1CountPass() common.TestRun {
	row1 := getPassedTest("TestSanitySha3_256")
	row2 := getPassedTest("TestSanitySha2_256")
	row3 := getPassedTest("TestSanitySha3_384")
	row4 := getPassedTest("TestSanitySha2_384")
	row5 := getPassedTest("TestSanityKmac128")
	row6 := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632497249121800000\n")
	row7 := getPassedTestElapsedOutput("TestSha3", 0.23, "0.23", "    hash_test.go:158: math rand seed is 1632497249122032000\n")
	row8 := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row9 := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")

	testRun := common.TestRun{
		Rows: []common.TestResultRow{
			row1,
			row2,
			row3,
			row4,
			row5,
			row6,
			row7,
			row8,
			row9,
		},
	}
	return testRun
}

func GetTestData_Level1_1Count1FailRestPass() common.TestRun {
	row1 := getFailedTest_TestSanitySha3_256()
	row2 := getPassedTest("TestSanitySha3_384")
	row3 := getPassedTest("TestSanitySha2_256")
	row4 := getPassedTest("TestSanitySha2_384")
	row5 := getPassedTest("TestSanityKmac128")
	row6 := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632498687765218000\n")
	row7 := getPassedTestElapsedOutput("TestSha3", 0.23, "0.23", "    hash_test.go:158: math rand seed is 1632498687765661000\n")
	row8 := getPassedTestElapsed("TestSha3/SHA3_256", 0.11, "0.11")
	row9 := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")

	testRun := common.TestRun{
		Rows: []common.TestResultRow{
			row1,
			row2,
			row3,
			row4,
			row5,
			row6,
			row7,
			row8,
			row9,
		},
	}
	return testRun
}

func GetTestData_Level1_1Count2SkippedRestPass() common.TestRun {
	row1 := getPassedTest("TestSanitySha3_256")
	row2 := getPassedTest("TestSanitySha3_384")
	row3 := getPassedTest("TestSanitySha2_384")
	row4 := getPassedTest("TestSanityKmac128")
	row5 := getPassedTestElapsedOutput("TestSha3", 0.24, "0.24", "    hash_test.go:160: math rand seed is 1633518697589650000\n")
	row6 := getPassedTestElapsed("TestSha3/SHA3_256", 0.11, "0.11")
	row7 := getPassedTestElapsed("TestSha3/SHA3_384", 0.13, "0.13")

	testRun := common.TestRun{
		Rows: []common.TestResultRow{
			row1,
			row2,
			row3,
			row4,
			row5,
			row6,
			row7,
		},
	}
	return testRun
}

func GetTestData_Level1_2CountPass() common.TestRun {
	row1 := getPassedTest("TestSanitySha3_256")
	row2 := getPassedTest("TestSanitySha2_256")
	row3 := getPassedTest("TestSanitySha3_384")
	row4 := getPassedTest("TestSanitySha2_384")
	row5 := getPassedTest("TestSanityKmac128")
	row6a := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1633358050203144000\n")
	row6b := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1633358050430256000\n")
	row7a := getPassedTestElapsedOutput("TestSha3", 0.23, "0.23", "    hash_test.go:158: math rand seed is 1633358050203374000\n")
	row7b := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1633358050430467000\n")

	row8 := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row9 := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")

	testRun := common.TestRun{
		Rows: []common.TestResultRow{
			row1,
			row2,
			row3,
			row4,
			row5,
			row6a,
			row7a,
			row8,
			row9,

			row1,
			row2,
			row3,
			row4,
			row5,
			row6b,
			row7b,
			row8,
			row9,
		},
	}
	return testRun
}

func GetTestData_Level1_10CountPass() common.TestRun {
	row1 := getPassedTest("TestSanitySha3_256")
	row2 := getPassedTest("TestSanitySha2_256")
	row3 := getPassedTest("TestSanitySha3_384")
	row4 := getPassedTest("TestSanitySha2_384")
	row5 := getPassedTest("TestSanityKmac128")

	row6a := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739552470379000\n")
	row6b := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739552696815000\n")
	row6c := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739552917474000\n")
	row6d := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739553140451000\n")
	row6e := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739553362249000\n")
	row6f := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739553605325000\n")
	row6g := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739553826502000\n")
	row6h := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739554054239000\n")
	row6i := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739554280043000\n")
	row6j := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739554500707000\n")

	row7a := getPassedTestElapsedOutput("TestSha3", 0.23, "0.23", "    hash_test.go:158: math rand seed is 1632739552470723000\n")
	row7b := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739552697024000\n")
	row7c := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739552917708000\n")
	row7d := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739553140702000\n")
	row7e := getPassedTestElapsedOutput("TestSha3", 0.24, "0.24", "    hash_test.go:158: math rand seed is 1632739553362497000\n")
	row7f := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739553605582000\n")
	row7g := getPassedTestElapsedOutput("TestSha3", 0.23, "0.23", "    hash_test.go:158: math rand seed is 1632739553826733000\n")
	row7h := getPassedTestElapsedOutput("TestSha3", 0.23, "0.23", "    hash_test.go:158: math rand seed is 1632739554054464000\n")
	row7i := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739554280256000\n")
	row7j := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739554500935000\n")

	row8a := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8b := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8c := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8d := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8e := getPassedTestElapsed("TestSha3/SHA3_256", 0.12, "0.12")
	row8f := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8g := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8h := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8i := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8j := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")

	row9a := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")
	row9b := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")
	row9c := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")
	row9d := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")
	row9e := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")
	row9f := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")
	row9g := getPassedTestElapsed("TestSha3/SHA3_384", 0.13, "0.13")
	row9h := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")
	row9i := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")
	row9j := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")

	var testResultRows []common.TestResultRow

	for i := 0; i < 10; i++ {
		testResultRows = append(testResultRows, row1)
		testResultRows = append(testResultRows, row2)
		testResultRows = append(testResultRows, row3)
		testResultRows = append(testResultRows, row4)
		testResultRows = append(testResultRows, row5)
	}

	testResultRows = append(testResultRows, row6a)
	testResultRows = append(testResultRows, row6b)
	testResultRows = append(testResultRows, row6c)
	testResultRows = append(testResultRows, row6d)
	testResultRows = append(testResultRows, row6e)
	testResultRows = append(testResultRows, row6f)
	testResultRows = append(testResultRows, row6g)
	testResultRows = append(testResultRows, row6h)
	testResultRows = append(testResultRows, row6i)
	testResultRows = append(testResultRows, row6j)

	testResultRows = append(testResultRows, row7a)
	testResultRows = append(testResultRows, row7b)
	testResultRows = append(testResultRows, row7c)
	testResultRows = append(testResultRows, row7d)
	testResultRows = append(testResultRows, row7e)
	testResultRows = append(testResultRows, row7f)
	testResultRows = append(testResultRows, row7g)
	testResultRows = append(testResultRows, row7h)
	testResultRows = append(testResultRows, row7i)
	testResultRows = append(testResultRows, row7j)

	testResultRows = append(testResultRows, row8a)
	testResultRows = append(testResultRows, row8b)
	testResultRows = append(testResultRows, row8c)
	testResultRows = append(testResultRows, row8d)
	testResultRows = append(testResultRows, row8e)
	testResultRows = append(testResultRows, row8f)
	testResultRows = append(testResultRows, row8g)
	testResultRows = append(testResultRows, row8h)
	testResultRows = append(testResultRows, row8i)
	testResultRows = append(testResultRows, row8j)

	testResultRows = append(testResultRows, row9a)
	testResultRows = append(testResultRows, row9b)
	testResultRows = append(testResultRows, row9c)
	testResultRows = append(testResultRows, row9d)
	testResultRows = append(testResultRows, row9e)
	testResultRows = append(testResultRows, row9f)
	testResultRows = append(testResultRows, row9g)
	testResultRows = append(testResultRows, row9h)
	testResultRows = append(testResultRows, row9i)
	testResultRows = append(testResultRows, row9j)

	testRun := common.TestRun{
		Rows: testResultRows,
	}
	return testRun
}

func GetTestData_Level1_10CountSomeFailures() common.TestRun {
	row1 := getPassedTest("TestSanitySha3_256")
	row2 := getFailedTest_TestSanitySha2_256()
	row3 := getPassedTest("TestSanitySha3_384")
	row4 := getPassedTest("TestSanitySha2_384")
	row5 := getPassedTest("TestSanityKmac128")

	row6a := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739682184421000\n")
	row6b := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739682415309000\n")
	row6c := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739682637108000\n")
	row6d := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739682857435000\n")
	row6e := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739683077064000\n")
	row6f := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739683297507000\n")
	row6g := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739683518492000\n")
	row6h := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739683740724000\n")
	row6i := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739683980033000\n")
	row6j := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632739684200452000\n")

	row7a := getPassedTestElapsedOutput("TestSha3", 0.23, "0.23", "    hash_test.go:158: math rand seed is 1632739682184858000\n")
	row7b := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739682415616000\n")
	row7c := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739682637311000\n")
	row7d := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739682857668000\n")
	row7e := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739683077268000\n")
	row7f := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739683297711000\n")
	row7g := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739683518781000\n")
	row7h := getPassedTestElapsedOutput("TestSha3", 0.24, "0.24", "    hash_test.go:158: math rand seed is 1632739683740970000\n")
	row7i := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739683980266000\n")
	row7j := getPassedTestElapsedOutput("TestSha3", 0.22, "0.22", "    hash_test.go:158: math rand seed is 1632739684200658000\n")

	row8a := getPassedTestElapsed("TestSha3/SHA3_256", 0.11, "0.11")
	row8b := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8c := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8d := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8e := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8f := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8g := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8h := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")
	row8i := getPassedTestElapsed("TestSha3/SHA3_256", 0.12, "0.12")
	row8j := getPassedTestElapsed("TestSha3/SHA3_256", 0.1, "0.10")

	row9 := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")

	var testResultRows []common.TestResultRow

	for i := 0; i < 10; i++ {
		testResultRows = append(testResultRows, row1)
		testResultRows = append(testResultRows, row2)
		testResultRows = append(testResultRows, row3)
		testResultRows = append(testResultRows, row4)
		testResultRows = append(testResultRows, row5)
		testResultRows = append(testResultRows, row9)
	}

	testResultRows = append(testResultRows, row6a)
	testResultRows = append(testResultRows, row6b)
	testResultRows = append(testResultRows, row6c)
	testResultRows = append(testResultRows, row6d)
	testResultRows = append(testResultRows, row6e)
	testResultRows = append(testResultRows, row6f)
	testResultRows = append(testResultRows, row6g)
	testResultRows = append(testResultRows, row6h)
	testResultRows = append(testResultRows, row6i)
	testResultRows = append(testResultRows, row6j)

	testResultRows = append(testResultRows, row7a)
	testResultRows = append(testResultRows, row7b)
	testResultRows = append(testResultRows, row7c)
	testResultRows = append(testResultRows, row7d)
	testResultRows = append(testResultRows, row7e)
	testResultRows = append(testResultRows, row7f)
	testResultRows = append(testResultRows, row7g)
	testResultRows = append(testResultRows, row7h)
	testResultRows = append(testResultRows, row7i)
	testResultRows = append(testResultRows, row7j)

	testResultRows = append(testResultRows, row8a)
	testResultRows = append(testResultRows, row8b)
	testResultRows = append(testResultRows, row8c)
	testResultRows = append(testResultRows, row8d)
	testResultRows = append(testResultRows, row8e)
	testResultRows = append(testResultRows, row8f)
	testResultRows = append(testResultRows, row8g)
	testResultRows = append(testResultRows, row8h)
	testResultRows = append(testResultRows, row8i)
	testResultRows = append(testResultRows, row8j)

	testRun := common.TestRun{
		Rows: testResultRows,
	}
	return testRun
}

func GetTestData_Level1_5CountSingleNilTest() common.TestRun {
	var testResultRows []common.TestResultRow
	row1 := getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack()

	for i := 0; i < 5; i++ {
		testResultRows = append(testResultRows, row1)
	}

	testRun := common.TestRun{
		Rows: testResultRows,
	}
	return testRun
}

func GetTestData_Level1_5CountMultipleNilTests() common.TestRun {
	var testResultRows []common.TestResultRow
	row1 := getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack()
	row2 := getPassedTestPackageElapsedOutput("TestEncodableRandomBeaconPrivKeyMsgPack", "github.com/onflow/flow-go/model/encodable", 0, "0.00", "    keys_test.go:245: bytes: 194\n")

	for i := 0; i < 4; i++ {
		testResultRows = append(testResultRows, row1)
	}
	testResultRows = append(testResultRows, row2)

	testRun := common.TestRun{
		Rows: testResultRows,
	}
	return testRun
}

func GetTestData_Leve1_3CountNilWithNormalTests() common.TestRun {
	var testResultRows []common.TestResultRow
	row1 := getPassedTestPackage("TestEncodableNetworkPrivKey", "github.com/onflow/flow-go/model/encodable")
	row2 := getPassedTestPackage("TestEncodableNetworkPrivKeyNil", "github.com/onflow/flow-go/model/encodable")
	row3 := getPassedTestPackage("TestEncodableNetworkPubKey", "github.com/onflow/flow-go/model/encodable")
	row4 := getPassedTestPackage("TestEncodableNetworkPubKeyNil", "github.com/onflow/flow-go/model/encodable")
	row5 := getPassedTestPackage("TestEncodableRandomBeaconPrivKey", "github.com/onflow/flow-go/model/encodable")
	row6 := getPassedTestPackage("TestEncodableRandomBeaconPrivKeyNil", "github.com/onflow/flow-go/model/encodable")
	row7 := getPassedTestPackage("TestEncodableRandomBeaconPubKey", "github.com/onflow/flow-go/model/encodable")
	row8 := getPassedTestPackage("TestEncodableRandomBeaconPubKeyNil", "github.com/onflow/flow-go/model/encodable")
	row9 := getPassedTestPackage("TestEncodableStakingPrivKey", "github.com/onflow/flow-go/model/encodable")
	row10 := getPassedTestPackage("TestEncodableStakingPrivKeyNil", "github.com/onflow/flow-go/model/encodable")
	row11 := getPassedTestPackage("TestEncodableStakingPubKey", "github.com/onflow/flow-go/model/encodable")
	row12 := getPassedTestPackage("TestEncodableStakingPubKeyNil", "github.com/onflow/flow-go/model/encodable")
	row13 := getPassedTestPackage("TestIsHexString", "github.com/onflow/flow-go/model/encodable")

	row14 := getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack()

	for i := 0; i < 3; i++ {
		testResultRows = append(testResultRows, row1)
		testResultRows = append(testResultRows, row2)
		testResultRows = append(testResultRows, row3)
		testResultRows = append(testResultRows, row4)
		testResultRows = append(testResultRows, row5)
		testResultRows = append(testResultRows, row6)
		testResultRows = append(testResultRows, row7)
		testResultRows = append(testResultRows, row8)
		testResultRows = append(testResultRows, row9)
		testResultRows = append(testResultRows, row10)
		testResultRows = append(testResultRows, row11)
		testResultRows = append(testResultRows, row12)
		testResultRows = append(testResultRows, row13)
		testResultRows = append(testResultRows, row14)
	}

	testRun := common.TestRun{
		Rows: testResultRows,
	}
	return testRun
}

// ************** Level 2 - Expected Test Runs Functions *****************
// following functions are used by unit tests for constructing expected TestRuns data

func GetTestData_Level2_1FailureRestPass() []common.TestRun {
	row1_1 := getFailedTest_TestSanitySha3_256()
	row2_1 := getPassedTest("TestSanitySha3_384")
	row3_1 := getPassedTest("TestSanitySha2_256")
	row4_1 := getPassedTest("TestSanitySha2_384")
	row5_1 := getPassedTest("TestSanityKmac128")
	row6_1 := getPassedTestElapsedOutput("TestHashersAPI", 0, "0.00", "    hash_test.go:114: math rand seed is 1632498687765218000\n")
	row7_1 := getPassedTestElapsedOutput("TestSha3", 0.23, "0.23", "    hash_test.go:158: math rand seed is 1632498687765661000\n")
	row8_1 := getPassedTestElapsed("TestSha3/SHA3_256", 0.11, "0.11")
	row9_1 := getPassedTestElapsed("TestSha3/SHA3_384", 0.12, "0.12")

	var testResult1Rows []common.TestResultRow
	testResult1Rows = append(testResult1Rows, row1_1)
	testResult1Rows = append(testResult1Rows, row2_1)
	testResult1Rows = append(testResult1Rows, row3_1)
	testResult1Rows = append(testResult1Rows, row4_1)
	testResult1Rows = append(testResult1Rows, row5_1)
	testResult1Rows = append(testResult1Rows, row6_1)
	testResult1Rows = append(testResult1Rows, row7_1)
	testResult1Rows = append(testResult1Rows, row8_1)
	testResult1Rows = append(testResult1Rows, row9_1)

	testRun1 := common.TestRun{
		Rows: testResult1Rows,
	}

	var testRuns []common.TestRun
	testRuns = append(testRuns, testRun1)
	return testRuns
}

func GetTestsData_Level2_1NoResultNoOtherTests() []common.TestRun {
	row1_1 := getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack()

	var testResult1Rows []common.TestResultRow
	testResult1Rows = append(testResult1Rows, row1_1)

	testRun1 := common.TestRun{
		Rows: testResult1Rows,
	}

	var testRuns []common.TestRun
	testRuns = append(testRuns, testRun1)
	return testRuns
}

func GetTestData_Level2_ManyNoResults() []common.TestRun {
	var testResult1Rows []common.TestResultRow
	row1_1 := getPassedTestPackage("TestEncodableNetworkPrivKey", "github.com/onflow/flow-go/model/encodable")
	row2_1 := getPassedTestPackage("TestEncodableNetworkPrivKeyNil", "github.com/onflow/flow-go/model/encodable")
	row3_1 := getPassedTestPackage("TestEncodableNetworkPubKey", "github.com/onflow/flow-go/model/encodable")
	row4_1 := getPassedTestPackage("TestEncodableNetworkPubKeyNil", "github.com/onflow/flow-go/model/encodable")
	row5_1 := getPassedTestPackage("TestEncodableRandomBeaconPrivKey", "github.com/onflow/flow-go/model/encodable")
	row6_1 := getPassedTestPackage("TestEncodableRandomBeaconPrivKeyNil", "github.com/onflow/flow-go/model/encodable")
	row7_1 := getPassedTestPackage("TestEncodableRandomBeaconPubKey", "github.com/onflow/flow-go/model/encodable")
	row8_1 := getPassedTestPackage("TestEncodableRandomBeaconPubKeyNil", "github.com/onflow/flow-go/model/encodable")
	row9_1 := getPassedTestPackage("TestEncodableStakingPrivKey", "github.com/onflow/flow-go/model/encodable")
	row10_1 := getPassedTestPackage("TestEncodableStakingPrivKeyNil", "github.com/onflow/flow-go/model/encodable")
	row11_1 := getPassedTestPackage("TestEncodableStakingPubKey", "github.com/onflow/flow-go/model/encodable")
	row12_1 := getPassedTestPackage("TestEncodableStakingPubKeyNil", "github.com/onflow/flow-go/model/encodable")
	row13_1 := getPassedTestPackage("TestIsHexString", "github.com/onflow/flow-go/model/encodable")
	row14_1 := getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack()

	for i := 0; i < 3; i++ {
		testResult1Rows = append(testResult1Rows, row1_1)
		testResult1Rows = append(testResult1Rows, row2_1)
		testResult1Rows = append(testResult1Rows, row3_1)
		testResult1Rows = append(testResult1Rows, row4_1)
		testResult1Rows = append(testResult1Rows, row5_1)
		testResult1Rows = append(testResult1Rows, row6_1)
		testResult1Rows = append(testResult1Rows, row7_1)
		testResult1Rows = append(testResult1Rows, row8_1)
		testResult1Rows = append(testResult1Rows, row9_1)
		testResult1Rows = append(testResult1Rows, row10_1)
		testResult1Rows = append(testResult1Rows, row11_1)
		testResult1Rows = append(testResult1Rows, row12_1)
		testResult1Rows = append(testResult1Rows, row13_1)
		testResult1Rows = append(testResult1Rows, row14_1)
	}

	testRun1 := common.TestRun{
		Rows: testResult1Rows,
	}

	row1_2 := getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack()

	var testResult2Rows []common.TestResultRow
	testResult2Rows = append(testResult2Rows, row1_2)
	testRun2 := common.TestRun{
		Rows: testResult2Rows,
	}

	row1_3 := getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack()
	row2_3 := getPassedTestPackageElapsedOutput("TestEncodableRandomBeaconPrivKeyMsgPack", "github.com/onflow/flow-go/model/encodable", 0, "0.00", "    keys_test.go:245: bytes: 194\n")

	var testResult3Rows []common.TestResultRow
	for i := 0; i < 4; i++ {
		testResult3Rows = append(testResult3Rows, row1_3)
	}
	testResult3Rows = append(testResult3Rows, row2_3)

	testRun3 := common.TestRun{
		Rows: testResult3Rows,
	}

	var testResult4Rows []common.TestResultRow
	row1_4 := getNilTest_TestEncodableRandomBeaconPrivKeyMsgPack()

	for i := 0; i < 5; i++ {
		testResult4Rows = append(testResult4Rows, row1_4)
	}

	testRun4 := common.TestRun{
		Rows: testResult4Rows,
	}

	var testRuns []common.TestRun
	testRuns = append(testRuns, testRun1)
	testRuns = append(testRuns, testRun2)
	testRuns = append(testRuns, testRun3)
	testRuns = append(testRuns, testRun4)
	return testRuns
}
