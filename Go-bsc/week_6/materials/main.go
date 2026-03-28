package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func osArgsDemo() {
	fmt.Printf("Количество аргументов: %d\n", len(os.Args)-1)
	fmt.Printf("Имя программы: %s\n", os.Args[0])

	if len(os.Args) > 1 {
		fmt.Println("Аргументы:")
		for i, arg := range os.Args[1:] {
			fmt.Printf("  [%d] %s\n", i, arg)
		}
	}
	fmt.Println()
}

func flagDemo() {
	fs := flag.NewFlagSet("demo", flag.ContinueOnError)
	name := fs.String("name", "Студент", "имя для приветствия")
	count := fs.Int("count", 1, "количество повторений")
	verbose := fs.Bool("verbose", false, "подробный вывод")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Ошибка парсинга: %v\n", err)
		return
	}

	if *verbose {
		fmt.Printf("Параметры: name=%s, count=%d\n", *name, *count)
	}

	for range *count {
		fmt.Printf("Привет, %s!\n", *name)
	}
	fmt.Println()
}

func flagArgsDemo() {
	fs := flag.NewFlagSet("fileProcessor", flag.ContinueOnError)
	prefix := fs.String("prefix", "", "префикс для файлов")

	args := []string{"-prefix=backup_", "file1.txt", "file2.txt", "file3.txt"}
	if err := fs.Parse(args); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	files := fs.Args()
	fmt.Printf("Обработка %d файлов:\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s%s\n", *prefix, file)
	}
	fmt.Println()
}

func fileCreateWriteDemo() {
	file, err := os.Create("demo_example.txt/demo_example2.txt")
	if err != nil {
		fmt.Printf("Ошибка создания файла: %v\n", err)
		// return
	}


	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Ошибка закрытия: %v\n", err)
		}
	}()

	content := "Демонстрация работы с файлами\nВторая строка текста\n"
	n, err := file.WriteString(content)
	if err != nil {
		fmt.Printf("Ошибка записи: %v\n", err)
		return
	}

	fmt.Printf("Записано %d байт в файл demo_example.txt\n", n)
	fmt.Println()
}

func fileReadDemo() {
	content, err := os.ReadFile("demo_example.txt")
	if err != nil {
		fmt.Printf("Ошибка чтения: %v\n", err)
		return
	}

	fmt.Printf("Содержимое файла:\n%s\n", content)
}

func fileReadBufferedDemo() {
	file, err := os.Open("demo_example.txt")
	if err != nil {
		fmt.Printf("Ошибка открытия: %v\n", err)
		return
	}
	defer file.Close()

	buffer := make([]byte, 16)
	chunk := 1

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Ошибка чтения: %v\n", err)
			return
		}

		fmt.Printf("Chunk %d (%d bytes): %q\n", chunk, n, buffer[:n])
		chunk++
	}
	fmt.Println()
}

func fileExistsDemo() {
	filename := "demo_example.txt"

	exists, err := fileExists(filename)
	if err != nil {
		fmt.Printf("Ошибка проверки: %v\n", err)
		return
	}

	if exists {
		fmt.Printf("Файл %s существует\n", filename)

		info, err := os.Stat(filename)
		if err == nil {
			fmt.Printf("  Размер: %d байт\n", info.Size())
			fmt.Printf("  Изменен: %v\n", info.ModTime().Format("2006-01-02 15:04:05"))
			fmt.Printf("  Директория: %v\n", info.IsDir())

		}
	} else {
		fmt.Printf("Файл %s не найден\n", filename)
	}
	fmt.Println()
}

func fileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}


func dirOperationsDemo() {
	if err := os.MkdirAll("demo_dirs/nested/deep", 0755); err != nil {
		fmt.Printf("Ошибка создания директорий: %v\n", err)
		return
	}
	fmt.Println("Создана структура: demo_dirs/nested/deep")

	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Printf("Ошибка чтения: %v\n", err)
		return
	}

	fmt.Println("\nСодержимое текущей директории (первые 5):")
	count := 0
	for _, entry := range entries {
		if count >= 5 {
			break
		}
		prefix := "[FILE]"
		if entry.IsDir() {
			prefix = "[DIR] "
		}
		fmt.Printf("  %s %s\n", prefix, entry.Name())
		count++
	}
	fmt.Println()
}

func walkDemo() {
	fmt.Println("Обход demo_dirs:")
	err := filepath.Walk("demo_dirs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			fmt.Printf("  [DIR]  %s\n", path)
		} else {
			fmt.Printf("  [FILE] %s (%d bytes)\n", path, info.Size())
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Ошибка обхода: %v\n", err)
	}

	fmt.Println("\nОбход demo_dirs с WalkDir:")
	err = filepath.WalkDir("demo_dirs", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			fmt.Printf("  [DIR]  %s\n", path)
		} else {
			info, err := d.Info()
			if err != nil {
				fmt.Printf("  [FILE] %s (ошибка получения размера)\n", path)
			} else {
				fmt.Printf("  [FILE] %s (%d bytes)\n", path, info.Size())
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Ошибка обхода WalkDir: %v\n", err)
	}
	fmt.Println()
}

func ioReaderDemo() {
	reader := strings.NewReader("Пример использования io.Reader")
	readFromReader(reader)
	fmt.Println()
	file, err := os.Open("demo_example.txt")
	if err != nil {
		fmt.Printf("Ошибка открытия файла: %v\n", err)
		return
	}
	defer file.Close()
	readFromReader(file)
	fmt.Println()
}

func readFromReader(reader io.Reader) {
	buffer := make([]byte, 10)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			break
		}

		fmt.Printf("Прочитано %d байт: %q\n", n, buffer[:n])
	}
	fmt.Println()
}

func ioCopyDemo() {
	source := strings.NewReader("Данные для копирования через io.Copy")

	dest, err := os.Create("demo_copy.txt")
	if err != nil {
		fmt.Printf("Ошибка создания: %v\n", err)
		return
	}
	defer dest.Close()

	written, err := io.Copy(dest, source)
	if err != nil {
		fmt.Printf("Ошибка копирования: %v\n", err)
		return
	}

	fmt.Printf("Скопировано %d байт в demo_copy.txt\n", written)
	fmt.Println()
}

func ioUtilityFunctionsDemo() {
	data := "Тестовые данные"
	reader := strings.NewReader(data)

	limited := io.LimitReader(reader, 10)
	result, _ := io.ReadAll(limited)
	fmt.Printf("LimitReader (10 байт): %q\n", result)

	reader1 := strings.NewReader("Часть 1. ")
	reader2 := strings.NewReader("Часть 2.")
	multi := io.MultiReader(reader1, reader2)
	combined, _ := io.ReadAll(multi)
	fmt.Printf("MultiReader: %q\n", combined)

	var buf1, buf2 strings.Builder
	writer := io.MultiWriter(&buf1, &buf2)
	io.WriteString(writer, "Запись в два буфера")
	fmt.Printf("MultiWriter buf1: %q\n", buf1.String())
	fmt.Printf("MultiWriter buf2: %q\n", buf2.String())

	fmt.Println()
}

func stringsSearchDemo() {
	text := "Язык программирования Go создан в Google"

	fmt.Printf("Текст: %s\n\n", text)
	fmt.Printf("Contains 'Go': %v\n", strings.Contains(text, "Go"))
	fmt.Printf("HasPrefix 'Язык': %v\n", strings.HasPrefix(text, "Язык"))
	fmt.Printf("HasSuffix 'Google': %v\n", strings.HasSuffix(text, "Google"))
	fmt.Printf("Count 'о': %d\n", strings.Count(text, "о"))
	fmt.Printf("Index 'Go': %d\n", strings.Index(text, "Go"))
	fmt.Printf("LastIndex 'о': %d\n", strings.LastIndex(text, "о"))

	fmt.Println()
}

func stringsTransformDemo() {
	original := "  Golang Programming  "

	fmt.Printf("Исходная: '%s'\n", original)
	fmt.Printf("Upper: '%s'\n", strings.ToUpper(original))
	fmt.Printf("Lower: '%s'\n", strings.ToLower(original))
	fmt.Printf("TrimSpace: '%s'\n", strings.TrimSpace(original))
	fmt.Printf("Trim '[]': '%s'\n", strings.Trim("[[Golang]]", "[]"))
	fmt.Printf("Replace 'Programming'->'Development': '%s'\n",
		strings.Replace(original, "Programming", "Development", 1))
	fmt.Printf("ReplaceAll ' '->'_': '%s'\n", strings.ReplaceAll(original, " ", "_"))

	fmt.Println()
}

func stringsSplitJoinDemo() {
	csv := "яблоко,банан,апельсин,груша"
	fruits := strings.Split(csv, ",")

	fmt.Printf("CSV: %s\n", csv)
	fmt.Printf("Split: %v\n", fruits)

	limited := strings.SplitN("a-b-c-d-e", "-", 3)
	fmt.Printf("SplitN (n=3): %v\n", limited)

	text := "  слово1   слово2\tслово3\nслово4  "
	words := strings.Fields(text)
	fmt.Printf("Fields: %v\n", words)

	joined := strings.Join(fruits, " | ")
	fmt.Printf("Join: %s\n", joined)

	repeated := strings.Repeat("Go! ", 3)
	fmt.Printf("Repeat: %s\n", repeated)

	fmt.Println()
}

func stringsBuilderDemo() {
	var builder strings.Builder
	builder.Grow(100)

	words := []string{"strings", "Builder", "эффективнее", "конкатенации"}

	for i, word := range words {
		builder.WriteString(word)
		if i < len(words)-1 {
			builder.WriteString(" ")
		}
	}

	fmt.Printf("Результат: %s\n", builder.String())
	fmt.Printf("Длина: %d, Capacity: %d\n", builder.Len(), builder.Cap())

	builder.Reset()
	builder.WriteString("После Reset")
	fmt.Printf("После Reset: %s\n", builder.String())

	fmt.Println()
}

func slicesSearchDemo() {
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}
	names := []string{"Алиса", "Боб", "Виктор"}

	fmt.Printf("numbers: %v\n", numbers)
	fmt.Printf("names: %v\n\n", names)

	fmt.Printf("Contains(numbers, 5): %v\n", slices.Contains(numbers, 5))
	fmt.Printf("Contains(names, 'Боб'): %v\n", slices.Contains(names, "Боб"))
	fmt.Printf("Index(numbers, 9): %d\n", slices.Index(numbers, 9))
	fmt.Printf("Max(numbers): %d\n", slices.Max(numbers))
	fmt.Printf("Min(numbers): %d\n", slices.Min(numbers))

	slice1 := []int{1, 2, 3}
	slice2 := []int{1, 2, 3}
	slice3 := []int{3, 2, 1}

	fmt.Printf("\nslice1==slice2: %v\n", slices.Equal(slice1, slice2))
	fmt.Printf("slice1==slice3: %v\n", slices.Equal(slice1, slice3))

	fmt.Println()
}

func slicesSortDemo() {
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}
	fmt.Printf("До сортировки: %v\n", numbers)

	slices.Sort(numbers)
	fmt.Printf("После Sort: %v\n", numbers)
	fmt.Printf("IsSorted: %v\n", slices.IsSorted(numbers))

	slices.Reverse(numbers)
	fmt.Printf("После Reverse: %v\n", numbers)

	words := []string{"банан", "яблоко", "апельсин"}
	sortedWords := slices.Clone(words)
	slices.Sort(sortedWords)

	fmt.Printf("\nИсходные: %v\n", words)
	fmt.Printf("Отсортированные: %v\n", sortedWords)

	fmt.Println()
}

func slicesModifyDemo() {
	numbers := []int{1, 2, 5, 6}
	fmt.Printf("Исходный: %v\n", numbers)

	numbers = slices.Insert(numbers, 2, 3, 4)
	fmt.Printf("Insert(2, 3, 4): %v\n", numbers)

	numbers = slices.Delete(numbers, 1, 3)
	fmt.Printf("Delete(1, 3): %v\n", numbers)

	duplicates := []int{1, 2, 2, 3, 3, 3, 4, 5, 5}
	fmt.Printf("\nС дубликатами: %v\n", duplicates)

	slices.Sort(duplicates)
	unique := slices.Compact(duplicates)
	fmt.Printf("После Compact: %v\n", unique)

	fmt.Println()
}

func slicesAdvancedDemo() {
	original := []int{1, 2, 3, 4, 5}
	cloned := slices.Clone(original)

	fmt.Printf("Оригинал: %v\n", original)
	fmt.Printf("Клон: %v\n", cloned)

	cloned[0] = 100
	fmt.Printf("После изменения клона:\n")
	fmt.Printf("  Оригинал: %v\n", original)
	fmt.Printf("  Клон: %v\n", cloned)

	var growing []int
	fmt.Printf("\nНачальная capacity: %d\n", cap(growing))

	growing = slices.Grow(growing, 100)
	fmt.Printf("После Grow(100): capacity=%d\n", cap(growing))

	longSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	shortSlice := slices.Clip(longSlice[:3])

	fmt.Printf("\nДлинный: len=%d cap=%d\n", len(longSlice), cap(longSlice))
	fmt.Printf("Короткий: len=%d cap=%d\n", len(shortSlice), cap(shortSlice))

	slice1 := []string{"apple", "banana"}
	slice2 := []string{"APPLE", "BANANA"}

	equal := slices.EqualFunc(slice1, slice2, func(a, b string) bool {
		return strings.EqualFold(a, b)
	})
	fmt.Printf("\nEqualFunc (без учета регистра): %v\n", equal)

	fmt.Println()
}

func prettyPrintMap[K comparable, V any](m map[K]V, title string) {
	if title != "" {
		fmt.Printf("%s:\n", title)
	}

	keys := slices.Collect(maps.Keys(m))
	slices.SortFunc(keys, func(a, b K) int {
		return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
	})

	for _, key := range keys {
		fmt.Printf("  %v: %v\n", key, m[key])
	}
}

func mapsBasicsDemo() {
	scores := map[string]int{
		"lisa":   95,
		"bob":    87,
		"victor": 92,
	}

	prettyPrintMap(scores, "Исходная мапа")

	backup := maps.Clone(scores)
	prettyPrintMap(backup, "\nКлон")

	scores["lisa"] = 100
	fmt.Println("\nПосле изменения оригинала:")
	prettyPrintMap(scores, "Оригинал")
	prettyPrintMap(backup, "Клон")

	other := map[string]int{"lisa": 95, "bob": 87, "victor": 92}

	fmt.Printf("\nEqual(backup, other): %v\n", maps.Equal(backup, other))
	fmt.Printf("Equal(scores, backup): %v\n", maps.Equal(scores, backup))

	fmt.Println()
}

func mapsKeysValuesDemo() {
	inventory := map[string]int{
		"apple":  50,
		"banana": 30,
		"orange": 25,
		"pear":   40,
	}

	prettyPrintMap(inventory, "Инвентарь")

	products := slices.Collect(maps.Keys(inventory))
	quantities := slices.Collect(maps.Values(inventory))

	fmt.Printf("\nТовары: %v\n", products)
	fmt.Printf("Количества: %v\n", quantities)

	slices.Sort(products)
	fmt.Printf("Товары (sorted): %v\n", products)

	fmt.Println("\nОтчет:")
	for _, product := range products {
		fmt.Printf("  - %s: %d шт.\n", product, inventory[product])
	}

	total := 0
	for _, q := range quantities {
		total += q
	}
	fmt.Printf("Всего: %d\n", total)

	fmt.Println()
}

func mapsMergeDemo() {
	defaults := map[string]string{
		"host":     "localhost",
		"port":     "8080",
		"timeout":  "30s",
		"loglevel": "info",
	}

	user := map[string]string{
		"host": "example.com",
		"port": "9000",
		"ssl":  "true",
	}

	prettyPrintMap(defaults, "Defaults")
	prettyPrintMap(user, "\nUser")

	final := maps.Clone(defaults)
	maps.Copy(final, user)

	prettyPrintMap(final, "\nFinal")

	fmt.Println()
}

func mapsFilterDemo() {
	data := map[string]interface{}{
		"name":   "lisa",
		"age":    30,
		"score":  95.5,
		"active": true,
		"level":  5,
	}

	prettyPrintMap(data, "Исходные данные")

	numericData := make(map[string]interface{})
	for key, value := range data {
		switch value.(type) {
		case int, float64, float32:
			numericData[key] = value
		}
	}

	prettyPrintMap(numericData, "\nТолько числовые")

	fmt.Println()
}

func cleanupDemo() {
	files := []string{"demo_example.txt", "demo_copy.txt"}
	for _, file := range files {
		if err := os.Remove(file); err == nil {
			fmt.Printf("Удален: %s\n", file)
		}
	}

	if err := os.RemoveAll("demo_dirs"); err == nil {
		fmt.Println("Удалена директория: demo_dirs")
	}

	fmt.Println()
}

func main() {
	// osArgsDemo()
	// flagDemo()
	// flagArgsDemo()

	// fileCreateWriteDemo()
	// fileReadDemo()
	// fileReadBufferedDemo()
	// fileExistsDemo()
	// dirOperationsDemo()
	// walkDemo()

	// ioReaderDemo()
	// ioCopyDemo()
	// ioUtilityFunctionsDemo()

	// stringsSearchDemo()
	// stringsTransformDemo()
	// stringsSplitJoinDemo()
	// stringsBuilderDemo()

	// slicesSearchDemo()
	// slicesSortDemo()
	// slicesModifyDemo()
	// slicesAdvancedDemo()

	// mapsBasicsDemo()
	// mapsKeysValuesDemo()
	// mapsMergeDemo()
	// mapsFilterDemo()

	// runCLIExample()
	// runLogAnalyzer()
	// demonstrateIOPipeline()
	// demonstrateDataTransformation()

	// cleanupDemo()
}
