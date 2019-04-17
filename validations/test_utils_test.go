// By naming this file with _test suffix it is not measured
// in the coverage report, although we do end-up with a strange file name...
package validate

import (
	. "github.com/onsi/gomega"
)

func assertNoValidationErrors(errors []YamlValidationIssue) {
	Ω(len(errors)).Should(Equal(0), "Validation issues detected: %v")
}

func expectSingleValidationError(actual []YamlValidationIssue, expectedMsg string, expectedLine int) {
	numOfErrors := len(actual)
	Ω(numOfErrors).Should(Equal(1), "A single validation issue expected but found: <%d>", numOfErrors)

	actualMsg := actual[0].Msg
	Ω(actualMsg).Should(Equal(expectedMsg), "Expecting <%s>.\n\t But found <%s>.", expectedMsg, actualMsg)
	Ω(actual[0].Line).Should(Equal(expectedLine))
}

func expectMultipleValidationError(actualIssues []YamlValidationIssue, expectedMsgs []string) {
	expectedNumOfErrors := len(expectedMsgs)
	actualNumOfErrors := len(actualIssues)
	Ω(actualNumOfErrors).Should(Equal(expectedNumOfErrors), "Wrong number of issues found expected <%d> but found: <%d>", expectedNumOfErrors, actualNumOfErrors)

	for _, issue := range actualIssues {
		Ω(expectedMsgs).Should(ContainElement(issue.Msg))
	}
}

func expectSingleSchemaIssue(actual []YamlValidationIssue, expectedMsg string) {
	numOfErrors := len(actual)
	Ω(numOfErrors).Should(Equal(1), "A single validation issue expected but found: <%d>", numOfErrors)

	Ω(actual[0].Msg).Should(Equal(expectedMsg))
}
