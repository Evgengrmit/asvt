package main

import (
	"fmt"
	"github.com/ernestosuarez/itertools"
	"math"
	"os"
	"strconv"
	"strings"
)

// Bit Представляет одну переменную в импликанте
type Bit int

// Переменная может:
const (
	Tilde Bit = -1 // Отсутствовать в импиканте
	False Bit = 0  // Быть инвертированной
	True  Bit = 1  // Быть прямой
)

// PrettyString Функция преобразования переменной в удобочитаемый вид
func (b Bit) PrettyString(index int) string {
	switch b {
	case Tilde:
		return ""
	case False:
		return "!x" + strconv.Itoa(index)

	case True:
		return "x" + strconv.Itoa(index)
	default:
		panic(fmt.Sprintf("bad bit value: %d", b))
	}
}

func (b Bit) String() string {
	switch b {
	case Tilde:
		return "~"
	case False:
		return "0"

	case True:
		return "1"
	default:
		panic(fmt.Sprintf("bad bit value: %d", b))
	}
}

// Представляет любую импликанту (терм) в виде набора переменных
type Term []Bit

// Функция преобразования импликанты в удобочитаемый вид
// Переводим каждую переменную в строку и конкатенируем их
// Стоит отметить, что переменные импликанты печатаются в обратном порядке:
// 01~~ -> "x1!x0"
func (a Term) PrettyString() string {
	var prettyString string
	for i, bit := range a {
		prettyString = bit.PrettyString(i) + prettyString
	}
	return prettyString
}

func (a Term) String() string {
	var prettyString string
	for _, bit := range a {
		prettyString = bit.String() + prettyString
	}
	return prettyString
}

// Функция сравнения импликант на равенство
func (a Term) Equals(b Term) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Функция нахождения расстояния между двумя импликантами
func (a Term) Distance(b Term) int {
	// Если импликанты зависят от разного количества
	// переменных, то считаем, что они не сравнимы
	if len(a) != len(b) {
		panic("expected terms lengths are equal")
	}
	dist := 0

	for i := range a {
		if a[i] != b[i] {
			dist++
		}
	}
	return dist
}

// Функция расчета веса импликанты
func (a Term) Weight() int {
	count := 0
	for _, bit := range a {
		if bit == True {
			count++
		}
	}
	return count
}

// Функция которая возвращает номер первой переменной,
// в которой импликанты различаются
func (a Term) DifferentBitIndex(b Term) int {
	// Если импликанты зависят от разного количества
	// переменных, то считаем, что они не сравнимы
	if len(a) != len(b) {
		panic("expected terms lengths are equal")
	}
	for i := range a {
		if a[i] != b[i] {
			return i
		}
	}
	return -1
}

// Функция проверки вхождения одной импликанты в другую
// К примеру импликанты 10~0 и 1~10 входят в импликанту 1~~0
func (a Term) Covers(b Term) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != Tilde {
			if a[i] != b[i] {
				return false
			}
		}
	}
	return true
}

// На первом шаге алгоритма мы разбиваем импликанты на группы по весу
// Данная структура представляет собой импликанту с отметкой о том,
// была ли она задействована при образовании новой склеенной импликанты
type GroupItem struct {
	Term
	IsGlued bool
}

// Вспомогательная функция для создания элемента по умолчанию
func NewGroupItem(term Term) GroupItem {
	return GroupItem{
		Term: term, IsGlued: false,
	}
}

// Представляет собой отображение веса на группу элементов с данным весом
type Groups map[int][]GroupItem

// Функция, которая создает группы, объединенные по весам
func GroupByWeight(terms []Term) Groups {
	if len(terms) == 0 {
		return nil
	}
	groups := make(map[int][]GroupItem, len(terms[0]))
	for _, term := range terms {
		// Считаем вес импликанты и добавляем ее в соответствующую группу
		weight := term.Weight()
		groups[weight] = append(groups[weight], NewGroupItem(term))
	}
	return groups
}

// Функцию реализует процесс склейки соседних по весу групп
// Результатом функции является набор импликант, образовавшихся при склеивании
func GlueGroups(a, b []GroupItem) (newTerms []Term) {
	// Перебираем все возможные пары элементов двух групп
	for i := range a {
		for j := range b {
			// Если он отличаются в одной позиции, то производим склейку
			if a[i].Distance(b[j].Term) == 1 {
				// Импликанты, которые участвовали в склеивании помечаются
				// поднятым флагом IsGlued для того, чтобы далее их можно было исключить
				a[i].IsGlued = true
				b[j].IsGlued = true
				// Создаем новую импликанту
				newTerm := make(Term, len(a[i].Term))
				copy(newTerm, a[i].Term)
				newTerm[a[i].DifferentBitIndex(b[j].Term)] = Tilde
				newTerms = append(newTerms, newTerm)

			}
		}
	}
	return newTerms
}

// Функция создает новый набор на основе входного, но без повторяющихся элементов
// К пр.: 10~1, 10~1, 010~, ~1~~ -> 10~1, 010~, ~1~~
func MakeUniqueSet(terms []Term) []Term {
	// Множество уникальных элементов
	uniqueSet := make([]Term, 0)
	for _, term := range terms {
		// Ищем элемент term во множестве уже найденных уникальных элементов
		found := false
		for _, termInSet := range uniqueSet {
			if termInSet.Equals(term) {
				found = true
				break
			}
		}
		// Если элемент еще не присутствует в уникальных, то добавляем его
		if !found {
			uniqueSet = append(uniqueSet, term)
		}
	}
	return uniqueSet
}

// Функцию реализует первый шаг алгоритма - склейка импликант
// Функция возвращает набор импликант, которые больше невозможно склеить
func Step1(impls []Term) []Term {
	// Формируем весовые группы
	groups := GroupByWeight(impls)
	// Склеиваем каждую весовую группу с предыдущей по весу, если такая имеется
	// Склеенные импликанты сохраняем
	glued := make([]Term, 0)
	for weight, groupB := range groups {
		if groupA, found := groups[weight-1]; found {
			newGlued := GlueGroups(groupA, groupB)
			glued = append(glued, newGlued...)
		}
	}
	// Если не произошло ни одного склеивания, то возвращаем входной набор импликант
	if len(glued) == 0 {
		return impls
	}
	// Ищем те импликанты, которые не были склеены
	unaffectedTerms := make([]Term, 0)
	for _, group := range groups {
		for _, term := range group {
			if !term.IsGlued {

				unaffectedTerms = append(unaffectedTerms, term.Term)
			}
		}
	}
	// К новым полученным импликантам добавляем те, что не были склеены
	impls = MakeUniqueSet(append(unaffectedTerms, glued...))
	// Запускаем следующий шаг рекурсии
	return Step1(impls)
}

// Линия представляет собой описание строки либо столбца таблицы для шагов 2, 3 и 4
// Каждой линии соответствует импликанта и служебная метка
type Line struct {
	Term     Term
	IsMarked bool
}

// Таблица для шагов 2, 3 и 4
// Содержит описания столбцов и строк, а так же таблицу меток о том, что
// импликанта строки покрывает импликанту столбца
type Table struct {
	Columns []Line
	Rows    []Line
	Marks   [][]bool
}

// Создает новую таблицу из наборов простых и исходных импликант
func NewTable(prime, source []Term) Table {
	t := Table{}
	// Каждой строке ставим соответствие простую импликанту
	for _, row := range prime {
		t.Rows = append(t.Rows, Line{
			Term: row, IsMarked: false,
		})
	}
	// Каждому столбцу ставим в соответствие исходную импликанту
	for _, column := range source {
		t.Columns = append(t.Columns, Line{Term: column, IsMarked: false})
	}
	// Заполняем таблицу меток
	t.Marks = make([][]bool, len(t.Rows))
	for i := range t.Rows {
		t.Marks[i] = make([]bool, len(t.Columns))
		for j := range t.Columns {
			// Если строка покрывает столбец, то ставим отметку
			if t.Rows[i].Term.Covers(t.Columns[j].Term) {
				t.Marks[i][j] = true

			}
		}
	}
	return t
}

func (t Table) PrettyString() string {
	const cellSize = 6
	var formatted string
	formatted += "|" + strings.Repeat(" ", cellSize) + "|"
	for _, column := range t.Columns {
		formatted += fmt.Sprintf("%"+strconv.Itoa(cellSize)+"s|", column.Term.String())
	}
	rowLen := len(formatted)
	formatLine := func(i, rowLen int) string {
		var formatted string
		underline := "\n" + strings.Repeat("-", rowLen) + "\n"
		if i == 0 {
			formatted += underline
		}

		formatted += fmt.Sprintf("|%"+strconv.Itoa(cellSize)+"s|", t.Rows[i].Term.String())
		for j := range t.Columns {
			var mark string
			if t.Marks[i][j] {
				mark = strings.Repeat("X", cellSize)

			} else {
				mark = "    "
			}

			formatted += fmt.Sprintf("%"+strconv.Itoa(cellSize)+"s|", mark)
		}
		formatted += underline
		return formatted
	}
	for i := range t.Rows {
		formatted += formatLine(i, rowLen)
	}
	return formatted
}

// Данная функция проверяет покрывает ли набор строк под номерами
// из rows все импликанты
func (t Table) IsRowsCovers(rows map[int]struct{}) bool {
	// Создаем массив меток о том, что столбцы были покрыты
	coveredColumns := make([]bool, len(t.Columns))
	// Проверяем каждую строку
	for i := range t.Rows {
		// Если номер строки не содержится в rowsTakeOff, то переходим к следующей строке
		if _, found := rows[i]; !found {
			continue
		}

		// Ставим метки о том, какие столбцы были покрыты
		for j := range t.Columns {
			if t.Marks[i][j] == true {
				coveredColumns[j] = true
			}
		}
	}
	// Если хоть один из столбцов не был покрыт, то набор оставшихся
	// строк не покрывает функцию полностью
	for _, covered := range coveredColumns {
		if !covered {
			return false
		}
	}
	return true
}

// Реализует 2, 3 и 4 шаги алгоритма
// Возвращает таблицу и набор существенных строк
func Steps2and3and4(prime, source []Term) (Table, map[int]struct{}) {
	// Создаем таблицу
	t := NewTable(prime, source)
	// Ищем существенные строки и столбцы
	for j := range t.Columns {
		marksInColumn := 0
		rowWithMark := 0
		// Считаем количество отметок в столбце
		for i := range t.Rows {
			if t.Marks[i][j] {
				marksInColumn++
				rowWithMark = i
			}
		}
		// Если в столбце только одна отметка, то строка и столбец,
		// соответствующие этой отметке существенны
		if marksInColumn == 1 {
			t.Columns[j].IsMarked = true
			t.Rows[rowWithMark].IsMarked = true
		}
	}
	// Составляем набор существенных строк
	essentials := make(map[int]struct{}, 0)
	for i := range t.Rows {
		if t.Rows[i].IsMarked {
			essentials[i] = struct{}{}
		}
	}
	return t, essentials
}

func GetCombinations(t Table, n int, essential map[int]struct{}) []map[int]struct{} {

	var essentials []int
	for rowIndex := range essential {
		essentials = append(essentials, rowIndex)
	}
	var indices []int
	for i := 0; i < len(t.Rows); i++ {
		if _, found := essential[i]; !found {
			indices = append(indices, i)
		}
	}

	combinations := make([]map[int]struct{}, 0)
	for indexCombination := range itertools.CombinationsInt(indices, n) {
		combination := make(map[int]struct{}, len(indexCombination))
		for _, index := range indexCombination {
			combination[index] = struct{}{}
		}
		for _, index := range essentials {
			combination[index] = struct{}{}
		}
		combinations = append(combinations, combination)
	}
	return combinations
}

// Функция реализует 5 шаг алгоритма
func Step5(t Table, essential map[int]struct{}) []Term {
	// Перебираем возможные комбинации строк таблицы начиная с 1 до len(t.Rows)
	var possibleResults [][]Term
	for i := 1; i < len(t.Rows)-len(essential); i++ {
		// Получаем все возможные комбинации из i строк таблицы
		combinations := GetCombinations(t, i, essential)
		fmt.Println("combinations from", i, "(", len(combinations), ")")
		// Проходим по каждой комбинации и проверяем, покрывает ли эта комбинация все строки
		for _, combination := range combinations {
			// Если покрывает, что возвращаем термы этой строки таблицы
			if t.IsRowsCovers(combination) {
				var result []Term
				for index := range combination {
					result = append(result, t.Rows[index].Term)
				}
				possibleResults = append(possibleResults, result)
			}
		}
	}
	// Если найти комбинацию не нашлось, то возвращаем все исходные термы
	var trivial []Term
	for _, row := range t.Rows {
		trivial = append(trivial, row.Term)
	}
	possibleResults = append(possibleResults, trivial)

	minComplexity := math.MaxInt32
	var minResult []Term
	for _, result := range possibleResults {
		complexity := 0
		for _, term := range result {
			complexity += len(term)
		}
		if complexity < minComplexity {
			minComplexity = complexity
			minResult = result
		}
	}
	return minResult
}

// Функция возвращает СДНФ от ФАЛ
func MakeSDNF(f []int) []Term {
	variableNumber := int(math.Log2(float64(len(f))))
	sdnf := make([]Term, 0, len(f))
	format := "%0" + strconv.Itoa(variableNumber) + "b"
	for i := range f {
		if f[i] == 1 {
			term := make(Term, 0, variableNumber)
			binaryString := fmt.Sprintf(format, i)
			for _, char := range binaryString {
				switch char {
				case '0':
					term = append(term, False)
				case '1':
					term = append(term, True)
				default:
					panic(fmt.Sprintf("unexpected char: %s", string(char)))
				}
			}
			sdnf = append(sdnf, term)
		}
	}
	return sdnf
}

// Функция форматирует импликанты в строку
// К пр.: [10~1, 010~, ~1~~] -> "x3!x1x0 + !x2x1!x0 + x3!x1x0"
func Format(impls []Term) string {
	var result string
	for i, impl := range impls {
		if i != 0 {
			result += " + "
		}
		result += impl.PrettyString()
	}

	return result
}

func PrettyString(terms []Term) string {
	var formatted string
	for i, impl := range terms {
		if i != 0 {
			formatted += " + "
		}
		formatted += impl.PrettyString()
	}
	return formatted
}

func String(terms []Term) string {
	var formatted string
	for i, impl := range terms {
		if i != 0 {
			formatted += ", "
		}
		formatted += impl.String()
	}
	return formatted
}

func main() {
	f := []int{1, 0, 1, 1, 1, 0, 1, 1, 0, 0,
		1, 1, 0, 0, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 0, 0, 0, 1, 0,
		1, 0, 1, 0, 0, 0, 1, 0, 0, 0,
		0, 0, 1, 0, 0, 0, 1, 0, 1, 0,
		0, 0, 1, 0, 0, 0, 1, 1, 1, 0,
		1, 0, 1, 0}
	//f := []int{
	//
	//	0, 0, 0, 1, 0, 0, 1, 0, 0, 1, // 00-09
	//	0, 1, 0, 1, 1, 1, 1, 0, 1, 1, // 10-19
	//	1, 0, 0, 0, 0, 0, 0, 1, 0, 0, // 20-29
	//	1, 1, 0, 1, 1, 1, 0, 0, 1, 0, // 30-39
	//	0, 1, 0, 1, 1, 1, 0, 1, 1, 0, // 40-49
	//	1, 1, 1, 0, 1, 1, 0, 1, 0, 0, // 50-59
	//	0, 1, 0, 1, // 60-63
	//}
	impls := MakeSDNF(f)
	fmt.Printf("source SDNF: %s\n", String(impls))

	primeImpls := Step1(impls)
	fmt.Printf("prime implicants: %s\n", String(primeImpls))

	table, essential := Steps2and3and4(primeImpls, impls)
	var coreImpls []Term
	for index := range essential {
		coreImpls = append(coreImpls, table.Rows[index].Term)
	}
	fmt.Printf("core implicants: %s\n", String(coreImpls))
	fmt.Println("table size after 4th step:", len(table.Rows))
	file, err := os.Create("./table.txt")
	if err != nil {
		panic(err)

	}
	defer file.Close()
	file.Write([]byte(table.PrettyString()))

	result := Step5(table, essential)
	// Сделаем проверку на то, что все исходные импликанты покрыты
	covered := 0
	total := 0
	for _, sourceImpl := range impls {
		isCover := false
		for _, minImpl := range result {
			if minImpl.Covers(sourceImpl) {
				isCover = true
			}
		}
		if !isCover {
			fmt.Print("E")

		} else {

		}
		total++
	}

	covered++
	fmt.Print(".")

	fmt.Printf("total coverage: %3.4f%%\n", float32(covered)/float32(total)*100)
	formatted := Format(result)
	count := 0
	for _, c := range formatted {
		if c == 'x' {
			count++
		}
	}
	fmt.Printf("result: %s\n", formatted)
	fmt.Printf("result complexity %d\n", count)
	fmt.Printf("implicants in result: %d\n", len(result))
}
