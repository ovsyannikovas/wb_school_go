package main

import "fmt"

// StovetopKettle - обычный чайник для плиты
type StovetopKettle struct {
	working bool
}

// PutOnStove - поставить на плиту
func (sk *StovetopKettle) PutOnStove() {
	sk.working = true
	fmt.Println("♨️  Чайник поставлен на плиту")
}

// TurnOffStove - снять с плиты
func (sk *StovetopKettle) TurnOffStove() {
	sk.working = false
	fmt.Println("🔥 Чайник снят с плиты")
}

// ElectricKettle - электрический чайник
type ElectricKettle struct {
	working bool
}

// TurnOnWithButton - включить кнопкой
func (ek *ElectricKettle) TurnOnWithButton() {
	ek.working = true
	fmt.Println("⚡ Электрический чайник включен кнопкой")
}

// TurnOffWithButton - выключить кнопкой
func (ek *ElectricKettle) TurnOffWithButton() {
	ek.working = false
	fmt.Println("🔌 Электрический чайник выключен кнопкой")
}

// KettleAdapter - общий интерфейс для всех чайников
type KettleAdapter interface {
	TurnOn()
	TurnOff()
}

// ElectricKettleAdapter - адаптер для электрического чайника
type ElectricKettleAdapter struct {
	*ElectricKettle
}

func (adapter *ElectricKettleAdapter) TurnOn() {
	adapter.TurnOnWithButton()
}

func (adapter *ElectricKettleAdapter) TurnOff() {
	adapter.TurnOffWithButton()
}

func NewElectricKettleAdapter(ek *ElectricKettle) KettleAdapter {
	return &ElectricKettleAdapter{ek}
}

// StovetopKettleAdapter - адаптер для обычного чайника
type StovetopKettleAdapter struct {
	*StovetopKettle
}

func (adapter *StovetopKettleAdapter) TurnOn() {
	adapter.PutOnStove()
}

func (adapter *StovetopKettleAdapter) TurnOff() {
	adapter.TurnOffStove()
}

func NewStovetopKettleAdapter(sk *StovetopKettle) KettleAdapter {
	return &StovetopKettleAdapter{sk}
}

// Функция для работы с любым чайником через адаптер
func useKettle(kettle KettleAdapter, name string) {
	fmt.Printf("\nИспользуем %s:\n", name)
	kettle.TurnOn()
	kettle.TurnOff()
}

func main() {
	fmt.Println("=== Паттерн Адаптер - простой пример ===")

	// Создаем чайники
	regularKettle := &StovetopKettle{}
	electricKettle := &ElectricKettle{}

	// Создаем адаптеры
	regularAdapter := NewStovetopKettleAdapter(regularKettle)
	electricAdapter := NewElectricKettleAdapter(electricKettle)

	// Используем чайники через единый интерфейс
	useKettle(regularAdapter, "обычный чайник")
	useKettle(electricAdapter, "электрический чайник")

	// Работа с массивом разных чайников
	fmt.Println("\n--- Все чайники в массиве ---")
	kettles := []KettleAdapter{regularAdapter, electricAdapter}

	for i, kettle := range kettles {
		fmt.Printf("\nЧайник %d:\n", i+1)
		kettle.TurnOn()
		kettle.TurnOff()
	}

	// Прямое использование
	fmt.Println("\n--- Прямое использование ---")
	fmt.Println("Обычный чайник:")
	regularAdapter.TurnOn()
	regularAdapter.TurnOff()

	fmt.Println("\nЭлектрический чайник:")
	electricAdapter.TurnOn()
	electricAdapter.TurnOff()
}
