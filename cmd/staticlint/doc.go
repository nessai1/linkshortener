/*
Package main. Линтер для приложения linkshortener.

# Линтер содержит в себе следующие анализаторы

## Стандартные анализаторы пакета passes

	printf - Анализатор, проверяющий верную сигнатуру строк с плейсхолдерами
	structtag - Анализатор, проверяющий корректность структурных тегов
	shadow - Анализатор, проверяющий потенциально неверное затенение переменных
	nilfunc - проверяющий бесполезное сравнение с nil
	loopclosure - Анализатор, проверяющий замыкания переменных цикла
	unreachable - Анализатор, проверяющий недоступные куски кода

## Анализаторы пакета staticcheck.io

	### Анализаторы класса SA, в который входит
		- неправильное использование стандартных библиотек;
		- проблемы с многопоточностью;
		- проблемы с тестами;
		- бесполезный код;
		- ошибочный код;
		- проблемы с производительностью;
		- сомнительные конструкции кода, c высокой вероятностью ошибочные;

	### Анализатор использования бесконечного for (S1006)

## Самописные анализаторы проекта, такие как
  - Анализатор exitcheck, проверяющий прямой вызов os.Exit в пакете main функции main

## Другие публичные анализаторы
  - Анализатор magic_numbers, проверяющий использование магических чисел в проекте
  - Анализатор containedctx, проверяющий объявление контекста в структурах
*/
package main