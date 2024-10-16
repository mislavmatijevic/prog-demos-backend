package lizard

import (
	"math"
	"os"
	"testing"
)

func TestCalculateScore_GivenBetterAndWorseCodeForSimpleTask_ScoreComparisonReportsHigherScoreForBetterCode(t *testing.T) {
	var betterCode = "#include<iostream>\n\nusing namespace std;\n\nint main() {\n        int N;\n    cout << \"Unesite pozitivan cijeli broj N: \";\n    cin >> N;\n\n    float rezultat = 0;\n    for (int i = 1; i <= N; i++) {\n        rezultat += (float)1/i;\n    }\n\n    cout << \"N-ti clan harmonijskog niza a_N = 1+1/2...+1/N iznosi: \" << rezultat << endl;\n\n    return 0;\n}\n"
	var worseCode = "#include<iostream>\n#include<iomanip>\n#include<stdexcept>\n#include<limits>\nusing namespace std;\nclass HarmonicSeries {\nprivate:\n    int N;             \n    float result;      \npublic:\n    \n    HarmonicSeries(int N) : N(N), result(0.0) {}\n    \n    void calculate() {\n        if (N <= 0) {\n            throw invalid_argument(\"N must be a positive integer.\");\n        }\n        for (int i = 1; i <= N; i++) {\n            result += static_cast<float>(1) / i;\n        }\n    }\n    \n    void displayResult() const {\n        cout << fixed << setprecision(5);\n        cout << \"N-ti clan harmonijskog niza a_N = 1+1/2...+1/N iznosi: \" << result << endl;\n    }\n};\nint getInput() {\n    int N;\n    while (true) {\n        cout << \"Unesite pozitivan cijeli broj N: \";\n        cin >> N;\n        \n        if (cin.fail() || N <= 0) {\n            cin.clear(); \n            cin.ignore(numeric_limits<streamsize>::max(), 'n'); \n            cout << \"Invalid input. Please enter a valid positive integer.\" << endl;\n        } else {\n            break; \n        }\n    }\n    return N;\n}\nint main() {\n    try {\n        \n        int N = getInput();\n        \n        HarmonicSeries harmonic(N);\n        harmonic.calculate();\n        \n        \n        harmonic.displayResult();\n    } catch (const exception& e) {\n        \n        cerr << \"Error: \" << e.what() << endl;\n    }\n    return 0;\n}\n"
	betterCppFile, _ := os.CreateTemp("", "solution_*.cpp")
	betterCppFile.Write([]byte(betterCode))
	worseCppFile, _ := os.CreateTemp("", "solution_*.cpp")
	worseCppFile.Write([]byte(worseCode))

	var betterScore, _ = CalculateScore(betterCppFile, 0)
	var worseScore, _ = CalculateScore(worseCppFile, 0)

	t.Logf("Better score: %v", betterScore)
	t.Logf("Worse score: %v", worseScore)

	if worseScore.HasBetterScoreThan(betterScore) {
		t.Fatalf("Score %v reported as better than %v!", worseCode, betterCode)
	}
}

func TestCalculateScore_GivenBetterAndWorseCodeForComplexTask_ScoreComparisonReportsHigherScoreForBetterCode(t *testing.T) {
	var betterCode = "#include <iostream>\n#include <vector>\nint sumArray(const std::vector<int>& arr) {int sum = 0;for (int num : arr) {sum += num;}return sum;}int main() {int array[] = {1, 2, 3, 4, 5};  // Example arraystd::vector<int> arr(array, array + 5);  // Construct vector explicitlystd::cout << \"Sum of elements: \" << sumArray(arr) << std::endl;return 0;}"
	var worseCode = "#include <iostream>\nclass ArraySum {private:int* arr;int size;public:ArraySum(int* array, int n) {arr = new int[n];size = n;for (int i = 0; i < n; i++) {arr[i] = array[i];}}~ArraySum() {delete[] arr;}int recursiveSum(int idx = 0) {if (idx >= size) {return 0;}if (idx % 2 == 0) {return arr[idx] + recursiveSum(idx + 1);} else {return recursiveSum(idx + 1);}}int sum() {int totalSum = 0;for (int i = 0; i < size; i++) {if (arr[i] >= 0) {totalSum += arr[i];}}return totalSum + recursiveSum();}};int main() {int arr[] = {1, 2, 3, 4, 5};ArraySum arraySum(arr, 5);std::cout << \"Sum of elements: \" << arraySum.sum() << std::endl;return 0;}"

	betterCppFile, _ := os.CreateTemp("", "solution_*.cpp")
	betterCppFile.Write([]byte(betterCode))
	worseCppFile, _ := os.CreateTemp("", "solution_*.cpp")
	worseCppFile.Write([]byte(worseCode))

	var betterScore, _ = CalculateScore(betterCppFile, 0)
	var worseScore, _ = CalculateScore(worseCppFile, 0)

	t.Logf("Better score: %v", betterScore)
	t.Logf("Worse score: %v", worseScore)

	if worseScore.HasBetterScoreThan(betterScore) {
		t.Fatalf("Score %v reported as better than %v!", worseCode, betterCode)
	}
}

func TestCalculateScore_GivenDifferentTaskComplexity_ScoreComparisonReportsHigherScoreForHigherTaskComplexity(t *testing.T) {
	var mockCode = "#include<iostream>\nusing namespace std;\nint main()\n{cout<<\"mock\";\nreturn 0;}"

	betterCppFile, _ := os.CreateTemp("", "solution_*.cpp")
	betterCppFile.Write([]byte(mockCode))
	worseCppFile, _ := os.CreateTemp("", "solution_*.cpp")
	worseCppFile.Write([]byte(mockCode))

	var invalidScore, _ = CalculateScore(betterCppFile, 0)
	var basicScore, _ = CalculateScore(betterCppFile, 1)
	var boostedScore, _ = CalculateScore(worseCppFile, 2)
	var evenMoreBoostedScore, _ = CalculateScore(worseCppFile, 3)

	t.Logf("Invalid score: %v", invalidScore)
	t.Logf("Basic Score: %v", basicScore)
	t.Logf("Boosted Score: %v", boostedScore)
	t.Logf("Even More Boosted Score: %v", evenMoreBoostedScore)

	if invalidScore.TotalScore != basicScore.TotalScore {
		t.Fatalf("Invalid score %v not equal to first complexity level score %v!", invalidScore, basicScore)
	}

	if basicScore.HasBetterScoreThan(boostedScore) {
		t.Fatalf("Complexity 1 score %v reported as better than complexity 2 score %v!", basicScore, boostedScore)
	}

	if boostedScore.HasBetterScoreThan(evenMoreBoostedScore) {
		t.Fatalf("Complexity 2 score %v reported as better than complexity 3 score %v!", boostedScore, evenMoreBoostedScore)
	}

	basicVsBoosted := boostedScore.TotalScore - basicScore.TotalScore
	if math.Round(float64(basicVsBoosted)) != 500 {
		t.Fatalf("Difference between the scores %v and %v is not expected 500, rather %v!", basicScore, boostedScore, basicVsBoosted)
	}

	boostedVsMoreBoosted := evenMoreBoostedScore.TotalScore - boostedScore.TotalScore
	if math.Round(float64(basicVsBoosted)) != math.Round(float64(boostedVsMoreBoosted)) {
		t.Fatalf("Difference between the task complexity levels score awards not same (%v != %v)!", basicVsBoosted, boostedVsMoreBoosted)
	}
}

func TestCalculateScore_GivenLizardKeywordComments_LizardGivesSameScore(t *testing.T) {
	var codeWithoutLizardComment = "#include<iostream>\n\nusing namespace std;\n\nint main() {\n        int N;\n    cout << \"Unesite pozitivan cijeli broj N: \";\n    cin >> N;\n\n    float rezultat = 0;\n    for (int i = 1; i <= N; i++) {\n        rezultat += (float)1/i;\n    }\n\n    cout << \"N-ti clan harmonijskog niza a_N = 1+1/2...+1/N iznosi: \" << rezultat << endl;\n\n    return 0;\n}\n"
	var codeWithCommentGenerated = "//GENERATED CODE\n" + codeWithoutLizardComment
	var codeWithCommentForgives = "#include<iostream>\n\nusing namespace std;\n\nint main() {// #lizard forgives\n        int N;\n    cout << \"Unesite pozitivan cijeli broj N: \";\n    cin >> N;\n\n    float rezultat = 0;\n    for (int i = 1; i <= N; i++) {\n        rezultat += (float)1/i;\n    }\n\n    cout << \"N-ti clan harmonijskog niza a_N = 1+1/2...+1/N iznosi: \" << rezultat << endl;\n\n    return 0;\n}\n"
	fileWithoutComment, _ := os.CreateTemp("", "solution_*.cpp")
	fileWithoutComment.Write([]byte(codeWithoutLizardComment))
	fileWithCommentGenerated, _ := os.CreateTemp("", "solution_*.cpp")
	fileWithCommentGenerated.Write([]byte(codeWithCommentGenerated))
	fileWithCommentForgives, _ := os.CreateTemp("", "solution_*.cpp")
	fileWithCommentForgives.Write([]byte(codeWithCommentForgives))

	var scoreWithoutComment, _ = CalculateScore(fileWithoutComment, 0)
	var scoreWithCommentGenerated, _ = CalculateScore(fileWithCommentGenerated, 0)
	var scoreWithCommentForgives, _ = CalculateScore(fileWithCommentForgives, 0)

	t.Logf("Score without comment: %v", scoreWithoutComment)
	t.Logf("Score with comment GENERATED CODE: %v", scoreWithCommentGenerated)
	t.Logf("Score with comment #lizard forgives: %v", scoreWithCommentForgives)

	if scoreWithoutComment.TotalScore != scoreWithCommentGenerated.TotalScore || scoreWithoutComment.TotalScore != scoreWithCommentForgives.TotalScore {
		t.Fatalf("Score %v different based on comment alone from one of: '%v' '%v'!", scoreWithoutComment, scoreWithCommentGenerated, scoreWithCommentForgives)
	}
}
