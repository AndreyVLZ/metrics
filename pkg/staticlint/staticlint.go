// staticlint состоит из:
// стандартных статических анализаторов пакета golang.org/x/tools/go/analysis/passes;
// всех анализаторов класса SA и QF  пакета staticcheck.io;
// публичных анализаторов: copyloopvar,
// собственного анализатора exitcheck.
package staticlint

import (
	"github.com/AndreyVLZ/metrics/pkg/staticlint/exitcheck"
	"github.com/karamaru-alpha/copyloopvar"
	"github.com/kkHAIKE/contextcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
)

func New() {
	analixers := []*analysis.Analyzer{
		appends.Analyzer,             // проверьте наличие пропущенных значений после добавления.
		asmdecl.Analyzer,             // сообщает о несоответствиях между файлами сборки и объявлениями Go.
		assign.Analyzer,              // обнаруживает бесполезные назначения.
		atomic.Analyzer,              // проверяет распространенные ошибки с помощью пакета sync/atomic.
		atomicalign.Analyzer,         // проверяет аргументы, не выровненные по 64 битам, для функций sync/atomic функций.
		bools.Analyzer,               // обнаруживает распространенные ошибки, связанные с логическими операторами.
		buildssa.Analyzer,            // создает SSA-представление безошибочного пакета и возвращает набор всех функций в нем.
		buildtag.Analyzer,            // проверяет теги сборки.
		cgocall.Analyzer,             // обнаруживает некоторые нарушения правил передачи указателя cgo.
		composite.Analyzer,           // проверяет составные литералы без ключей.
		copylock.Analyzer,            // проверяет блокировки, ошибочно переданные по значению.
		ctrlflow.Analyzer,            // предоставляет синтаксический граф потока управления (CFG) для тела функции.
		deepequalerrors.Analyzer,     // проверяет использование Reflection.DeepEqual со значениями ошибок.
		defers.Analyzer,              // проверяет распространенные ошибки в операторах defer.
		directive.Analyzer,           // проверяет известные директивы цепочки инструментов Go.
		errorsas.Analyzer,            // проверяет, что второй аргумент error.As является указателем на реализации типа error.
		fieldalignment.Analyzer,      // обнаруживает структуры, которые использовали бы меньше памяти, если бы их поля были отсортированы.
		findcall.Analyzer,            // служит тривиальным примером и тестом API анализа.
		framepointer.Analyzer,        // сообщает ассемблерный код, который затирает указатель кадра перед его сохранением.
		httpmux.Analyzer,             // отчет с использованием расширенных шаблонов ServeMux Go 1.22 в старых версиях Go.
		httpresponse.Analyzer,        // проверяет наличие ошибок с помощью ответов HTTP.
		ifaceassert.Analyzer,         // помечает невозможные утверждения типа интерфейса.
		inspect.Analyzer,             // предоставляет инспектор AST для синтаксических деревьев пакета.
		loopclosure.Analyzer,         // проверяет ссылки на включающие переменные цикла внутри вложенных функций.
		lostcancel.Analyzer,          // проверяет отсутствие вызова функции отмены контекста.
		nilfunc.Analyzer,             // проверяет бесполезные сравнения с nil.
		nilness.Analyzer,             // проверяет граф потока управления функции SSA и сообщает об ошибках, таких как разыменование нулевого указателя и вырожденные сравнения нулевого указателя.
		pkgfact.Analyzer,             // это демонстрация и проверка механизма фактов пакета.
		printf.Analyzer,              // проверяет согласованность строк и аргументов формата Printf.
		reflectvaluecompare.Analyzer, // проверяет случайное использование == или Reflection.DeepEqual для сравнения значений Reflection.Value.
		shadow.Analyzer,              // проверяет наличие затененных переменных.
		shift.Analyzer,               // проверяет сдвиги, превышающие ширину целого числа.
		sigchanyzer.Analyzer,         // обнаруживает неправильное использование небуферизованного сигнала в качестве аргумента signal.Notify.
		slog.Analyzer,                // проверяет наличие несовпадающих пар ключ-значение в вызовах log/slog.
		sortslice.Analyzer,           // проверяет вызовы sort.Slice, которые не используют тип среза в качестве первого аргумента.
		stdmethods.Analyzer,          // проверяет наличие орфографических ошибок в сигнатурах методов, аналогичных общеизвестным интерфейсам.
		stdversion.Analyzer,          // сообщает об использовании символов стандартной библиотеки, которые являются «слишком новыми» для версии Go, действующей в ссылающемся файле.
		stringintconv.Analyzer,       // помечает преобразования типов из целых чисел в строки.
		structtag.Analyzer,           // проверяет правильность формирования тегов полей структуры.
		testinggoroutine.Analyzer,    // обнаружениние вызовов Fatal из горутины тестирования.
		tests.Analyzer,               // проверяет распространенные ошибки использования тестов и примеров.
		timeformat.Analyzer,          // проверяет использование вызовов time.Format или time.Parse с неверным форматом.
		unmarshal.Analyzer,           // проверяет передачу типов, не являющихся указателями или неинтерфейсами, для функций демаршалинга и декодирования.
		unreachable.Analyzer,         // проверяет недоступный код.
		unsafeptr.Analyzer,           // проверяет недопустимые преобразования uintptr в unsafe.Pointer.
		unusedresult.Analyzer,        // проверяет неиспользуемые результаты вызовов определенных функций.
		unusedwrite.Analyzer,         // проверяет наличие неиспользуемых записей в элементы объекта структуры или массива.
		usesgenerics.Analyzer,        // проверяет использование универсальных функций, добавленных в Go 1.18.
		copyloopvar.NewAnalyzer(),    // определяет места копирования переменных цикла.
		contextcheck.NewAnalyzer(contextcheck.Configuration{DisableFact: true}), // проверка использует ли функция ненаследуемый контекст, что приведет к неработающей ссылке вызова.
		exitcheck.NewAnalyzer(), // проверяет прямой вызов os.Exit в функции main.
	}

	// определяем map подключаемых правил для staticcheck.
	checks := map[string]bool{
		"SA": true,
		"QF": true,
	}

	// cтатический анализ, находит ошибки и проблемы с производительностью
	for _, v := range staticcheck.Analyzers {
		// добавляем в массив нужные проверки
		if checks[v.Analyzer.Name] {
			analixers = append(analixers, v.Analyzer)
		}
	}

	multichecker.Main(analixers...)
}
