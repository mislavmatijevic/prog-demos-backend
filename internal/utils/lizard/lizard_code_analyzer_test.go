package lizard

import (
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

	var betterScore, _ = CalculateScore(betterCppFile)
	var worseScore, _ = CalculateScore(worseCppFile)

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

	var betterScore, _ = CalculateScore(betterCppFile)
	var worseScore, _ = CalculateScore(worseCppFile)

	t.Logf("Better score: %v", betterScore)
	t.Logf("Worse score: %v", worseScore)

	if worseScore.HasBetterScoreThan(betterScore) {
		t.Fatalf("Score %v reported as better than %v!", worseCode, betterCode)
	}
}
