package main

import (
	"container/list"
	"fmt"
	"sync"
)

// entry представляет элемент кэша: ключ и значение
type entry struct {
	key   string
	value interface{}
}

// LRUCache – реализация LRU-кэша с использованием container/list
type LRUCache struct {
	capacity int                      // максимальное количество элементов
	items    map[string]*list.Element // отображение ключа на элемент списка
	order    *list.List               // двусвязный список (самые свежие в начале, самые старые в конце)
	mu       sync.RWMutex             // для потокобезопасности (опционально)
}

// NewLRUCache создаёт новый кэш с заданной ёмкостью
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Get возвращает значение по ключу и булевый флаг, найден ли ключ.
// При обращении элемент перемещается в начало списка (самый свежий).
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if elem, ok := c.items[key]; ok {
		// перемещаем в начало
		c.order.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return nil, false
}

// Put добавляет или обновляет значение по ключу.
// Если кэш заполнен, удаляется самый старый элемент (в конце списка).
func (c *LRUCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		// обновляем значение и перемещаем в начало
		elem.Value.(*entry).value = value
		c.order.MoveToFront(elem)
		return
	}

	// если достигнут лимит, удаляем самый старый
	if c.order.Len() >= c.capacity {
		c.removeOldest()
	}

	// создаём новый элемент списка
	ent := &entry{key: key, value: value}
	elem := c.order.PushFront(ent)
	c.items[key] = elem
}

// removeOldest удаляет самый старый элемент (в конце списка)
func (c *LRUCache) removeOldest() {
	elem := c.order.Back()
	if elem != nil {
		c.order.Remove(elem)
		delete(c.items, elem.Value.(*entry).key)
	}
}

// Len возвращает текущее количество элементов в кэше
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// Print выводит содержимое кэша в порядке от самого свежего к самому старому
func (c *LRUCache) Print() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	fmt.Print("LRU Cache (newest → oldest): ")
	for e := c.order.Front(); e != nil; e = e.Next() {
		ent := e.Value.(*entry)
		fmt.Printf("[%s:%v] ", ent.key, ent.value)
	}
	fmt.Println()
}

func main() {
	// Создаём кэш на 3 элемента
	cache := NewLRUCache(3)

	cache.Put("a", 1)
	cache.Put("b", 2)
	cache.Put("c", 3)
	cache.Print() // a:1 b:2 c:3

	// Обращаемся к "a" – она становится самой свежей
	cache.Get("a")
	cache.Print() // a:1 c:3 b:2 (порядок изменился)

	// Добавляем новый элемент – вытесняется самый старый ("b")
	cache.Put("d", 4)
	cache.Print() // d:4 a:1 c:3

	// Проверяем наличие
	if val, ok := cache.Get("b"); !ok {
		fmt.Println("Key 'b' not found (evicted)") // должно быть выведено
	} else {
		fmt.Printf("b = %v\n", val)
	}
}
