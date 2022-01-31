package faker

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"pkg.re/essentialkaos/translit.v2"
)

var companyPrefix = []string{"ООО", "ОАО"}

var companySuffix = []string{
	"Авалон", "Аквилон", "Амазон", "Прогресс", "Торг", "Трейд",
	"Инвест", "Премьер", "Интер", "Скай", "Софт", "Хауз",
}

var companyPostfix = []string{
	"лаб", "эдванс", "про", "связь", "сейв", "партнер", "сервис",
	"майнинг", "дизайн", "креатив",
}

func getCompanyName() (string, string) {

	cp := companyPrefix[randomInt(len(companyPrefix))]
	cs := companySuffix[randomInt(len(companySuffix))]
	cpx := companyPostfix[randomInt(len(companyPostfix))]

	name := fmt.Sprintf("%s%s", cs, cpx)
	fullName := fmt.Sprintf("%s \"%s%s\"", cp, cs, cpx)
	return name, fullName
}

var countryList = []string{
	"Россия",
}

var cityList = []string{
	"Санк-Петербург", "Москва",
}

var indexList = []string{
	"191186", "129223",
}

var srteetList = []string{
	"Советская", "Молодежная", "Центральная", "Школьная", "Новая", "Садовая", "Лесная", "Смоленская",
	"Грина", "Крузенштерна", "Вознесенская",
}

var srteetPrefixList = []string{
	"пл.", "ул.", "наб.",
}

var maleFirstNameList = []string{
	"Александр", "Алексей", "Альберт", "Анатолий", "Андрей", "Антон", "Аркадий", "Арсений", "Артём",
	"Борис", "Вадим", "Валентин", "Валерий", "Василий", "Виктор", "Виталий", "Владимир", "Владислав",
	"Вячеслав", "Геннадий", "Георгий", "Герман", "Григорий", "Даниил", "Денис", "Дмитрий", "Евгений",
	"Егор", "Иван", "Игнатий", "Игорь", "Илья", "Константин", "Лаврентий", "Леонид", "Лука", "Макар",
	"Максим", "Матвей", "Михаил", "Никита", "Николай", "Олег", "Роман", "Семён", "Сергей", "Станислав",
	"Степан", "Фёдор", "Эдуард", "Юрий", "Ярослав",
}

var maleLastNameList = []string{
	"Смирнов", "Иванов", "Кузнецов", "Попов", "Соколов", "Лебедев", "Козлов", "Новиков", "Морозов", "Петров",
	"Волков", "Соловьев", "Васильев", "Зайцев", "Павлов", "Семенов", "Голубев", "Виноградов", "Богданов", "Воробьев",
	"Федоров", "Михайлов", "Беляев", "Тарасов", "Белов", "Комаров", "Орлов", "Киселев", "Макаров", "Андреев", "Ковалев",
	"Ильин", "Гусев", "Титов", "Кузьмин", "Кудрявцев", "Баранов", "Куликов", "Алексеев", "Степанов", "Яковлев", "Сорокин",
	"Сергеев", "Романов", "Захаров", "Борисов", "Королев", "Герасимов", "Пономарев", "Григорьев", "Лазарев", "Медведев",
	"Ершов", "Никитин", "Соболев", "Рябов", "Поляков", "Цветков", "Данилов", "Жуков", "Фролов", "Журавлев", "Николаев",
	"Крылов", "Максимов", "Сидоров", "Осипов", "Белоусов", "Федотов", "Дорофеев", "Егоров", "Матвеев", "Бобров", "Дмитриев",
	"Калинин", "Анисимов", "Петухов", "Антонов", "Тимофеев", "Никифоров", "Веселов", "Филиппов", "Марков", "Большаков",
	"Суханов", "Миронов", "Ширяев", "Александров", "Коновалов", "Шестаков", "Казаков", "Ефимов", "Денисов", "Громов", "Фомин",
	"Давыдов", "Мельников", "Щербаков", "Блинов", "Колесников", "Карпов", "Афанасьев", "Власов", "Маслов", "Исаков", "Тихонов",
	"Аксенов", "Гаврилов", "Родионов", "Котов", "Горбунов", "Кудряшов", "Быков", "Зуев", "Третьяков", "Савельев", "Панов",
	"Рыбаков", "Суворов", "Абрамов", "Воронов", "Мухин", "Архипов", "Трофимов", "Мартынов", "Емельянов", "Горшков", "Чернов",
	"Овчинников", "Селезнев", "Панфилов", "Копылов", "Михеев", "Галкин", "Назаров", "Лобанов", "Лукин", "Беляков", "Потапов",
	"Некрасов", "Хохлов", "Жданов", "Наумов", "Шилов", "Воронцов", "Ермаков", "Дроздов", "Игнатьев", "Савин", "Логинов",
	"Сафонов", "Капустин", "Кириллов", "Моисеев", "Елисеев", "Кошелев", "Костин", "Горбачев", "Орехов", "Ефремов", "Исаев",
	"Евдокимов", "Калашников", "Кабанов", "Носков", "Юдин", "Кулагин", "Лапин", "Прохоров", "Нестеров", "Харитонов",
	"Агафонов", "Муравьев", "Ларионов", "Федосеев", "Зимин", "Пахомов", "Шубин", "Игнатов", "Филатов", "Крюков", "Рогов",
	"Кулаков", "Терентьев", "Молчанов", "Владимиров", "Артемьев", "Гурьев", "Зиновьев", "Гришин", "Кононов", "Дементьев",
	"Ситников", "Симонов", "Мишин", "Фадеев", "Комиссаров", "Мамонтов", "Носов", "Гуляев", "Шаров", "Устинов", "Вишняков",
	"Евсеев", "Лаврентьев", "Брагин", "Константинов", "Корнилов", "Авдеев", "Зыков", "Бирюков", "Шарапов", "Никонов",
	"Щукин", "Дьячков", "Одинцов", "Сазонов", "Якушев", "Красильников", "Гордеев", "Самойлов", "Князев", "Беспалов",
	"Уваров", "Шашков", "Бобылев", "Доронин", "Белозеров", "Рожков", "Самсонов", "Мясников", "Лихачев", "Буров", "Сысоев",
	"Фомичев", "Русаков", "Стрелков", "Гущин", "Тетерин", "Колобов", "Субботин", "Фокин", "Блохин", "Селиверстов", "Пестов",
	"Кондратьев", "Силин", "Меркушев", "Лыткин", "Туров"}

var femaleFirstNameList = []string{"Анна", "Алёна", "Алевтина", "Александра", "Алина", "Алла",
	"Анастасия", "Ангелина", "Анжела", "Анжелика", "Антонида", "Антонина", "Анфиса", "Арина",
	"Валентина", "Валерия", "Варвара", "Василиса", "Вера", "Вероника", "Виктория", "Галина",
	"Дарья", "Евгения", "Екатерина", "Елена", "Елизавета", "Жанна", "Зинаида", "Зоя", "Ирина",
	"Кира", "Клавдия", "Ксения", "Лариса", "Лидия", "Любовь", "Людмила", "Маргарита", "Марина",
	"Мария", "Надежда", "Наталья", "Нина", "Оксана", "Ольга", "Раиса", "Регина", "Римма", "Светлана",
	"София", "Таисия", "Тамара", "Татьяна", "Ульяна", "Юлия",
}

var femaleLastNameList = []string{"Смирнова", "Иванова", "Кузнецова", "Попова", "Соколова", "Лебедева",
	"Козлова", "Новикова", "Морозова", "Петрова", "Волкова", "Соловьева", "Васильева", "Зайцева", "Павлова",
	"Семенова", "Голубева", "Виноградова", "Богданова", "Воробьева", "Федорова", "Михайлова", "Беляева",
	"Тарасова", "Белова", "Комарова", "Орлова", "Киселева", "Макарова", "Андреева", "Ковалева", "Ильина",
	"Гусева", "Титова", "Кузьмина", "Кудрявцева", "Баранова", "Куликова", "Алексеева", "Степанова",
	"Яковлева", "Сорокина", "Сергеева", "Романова", "Захарова", "Борисова", "Королева", "Герасимова",
	"Пономарева", "Григорьева", "Лазарева", "Медведева", "Ершова", "Никитина", "Соболева", "Рябова",
	"Полякова", "Цветкова", "Данилова", "Жукова", "Фролова", "Журавлева", "Николаева", "Крылова",
	"Максимова", "Сидорова", "Осипова", "Белоусова", "Федотова", "Дорофеева", "Егорова", "Матвеева",
	"Боброва", "Дмитриева", "Калинина", "Анисимова", "Петухова", "Антонова", "Тимофеева", "Никифорова",
	"Веселова", "Филиппова", "Маркова", "Большакова", "Суханова", "Миронова", "Ширяева", "Александрова",
	"Коновалова", "Шестакова", "Казакова", "Ефимова", "Денисова", "Громова", "Фомина", "Давыдова",
	"Мельникова", "Щербакова", "Блинова", "Колесникова", "Карпова", "Афанасьева", "Власова", "Маслова",
	"Исакова", "Тихонова", "Аксенова", "Гаврилова", "Родионова", "Котова", "Горбунова", "Кудряшова",
	"Быкова", "Зуева", "Третьякова", "Савельева", "Панова", "Рыбакова", "Суворова", "Абрамова", "Воронова",
	"Мухина", "Архипова", "Трофимова", "Мартынова", "Емельянова", "Горшкова", "Чернова", "Овчинникова",
	"Селезнева", "Панфилова", "Копылова", "Михеева", "Галкина", "Назарова", "Лобанова", "Лукина",
	"Белякова", "Потапова", "Некрасова", "Хохлова", "Жданова", "Наумова", "Шилова", "Воронцова",
	"Ермакова", "Дроздова", "Игнатьева", "Савина", "Логинова", "Сафонова", "Капустина", "Кириллова",
	"Моисеева", "Елисеева", "Кошелева", "Костина", "Горбачева", "Орехова", "Ефремова", "Исаева",
	"Евдокимова", "Калашникова", "Кабанова", "Носкова", "Юдина", "Кулагина", "Лапина", "Прохорова",
	"Нестерова", "Харитонова", "Агафонова", "Муравьева", "Ларионова", "Федосеева", "Зимина", "Пахомова",
	"Шубина", "Игнатова", "Филатова", "Крюкова", "Рогова", "Кулакова", "Терентьева", "Молчанова",
	"Владимирова", "Артемьева", "Гурьева", "Зиновьева", "Гришина", "Кононова", "Дементьева", "Ситникова",
	"Симонова", "Мишина", "Фадеева", "Комиссарова", "Мамонтова", "Носова", "Гуляева", "Шарова", "Устинова",
	"Вишнякова", "Евсеева", "Лаврентьева", "Брагина", "Константинова", "Корнилова", "Авдеева", "Зыкова",
	"Бирюкова", "Шарапова", "Никонова", "Щукина", "Дьячкова", "Одинцова", "Сазонова", "Якушева",
	"Красильникова", "Гордеева", "Самойлова", "Князева", "Беспалова", "Уварова", "Шашкова", "Бобылева",
	"Доронина", "Белозерова", "Рожкова", "Самсонова", "Мясникова", "Лихачева", "Бурова", "Сысоева",
	"Фомичева", "Русакова", "Стрелкова", "Гущина", "Тетерина", "Колобова", "Субботина", "Фокина", "Блохина",
	"Селиверстова", "Пестова", "Кондратьева", "Силина", "Меркушева", "Лыткина", "Турова",
}

var freeEmailList = []string{"yandex.ru", "ya.ru", "mail.ru", "gmail.com", "yahoo.com", "hotmail.com", "me.com"}

var titleList = []string{"менеджер", "наладчик", "помошник руководителя", "начальник отдела", "инженер", "сметчик",
	"проектировщик",
}

var nounList = []string{"Замена", "Неисправность", "Сбой", "Возгорание", "Тест", "Проверка работоспособности", "Обновление микропрошивки"}
var productList = []string{"Diode", "LED", "Rectifier", "Transistor", "JFET", "MOSFET", "Integrated Circuit", "LCD", "Cathode Ray Tube", "Vacuum Tube", "Battery", "Fuel Cell", "Power Supply"}

func randomInt(i int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(i)
}

// RandomInt Get three parameters , only first mandatory and the rest are optional
// (minimum_int, maximum_int, count)
// 		If only set one parameter :  An integer greater than minimum_int will be returned
// 		If only set two parameters : All integers between minimum_int and maximum_int will be returned, in a random order.
// 		If three parameters: `count` integers between minimum_int and maximum_int will be returned.
func RandomInt(parameters ...int) (p []int, err error) {
	switch len(parameters) {
	case 1:
		minInt := parameters[0]
		p = rand.Perm(minInt)
		for i := range p {
			p[i] += minInt
		}
	case 2:
		minInt, maxInt := parameters[0], parameters[1]
		p = rand.Perm(maxInt - minInt + 1)

		for i := range p {
			p[i] += minInt
		}
	case 3:
		minInt, maxInt := parameters[0], parameters[1]
		count := parameters[2]
		p = rand.Perm(maxInt - minInt + 1)

		for i := range p {
			p[i] += minInt
		}
		p = p[0:count]
	default:
		err = fmt.Errorf("error", len(parameters))
	}
	return p, err
}

// IntToString Convert slice int to slice string
func IntToString(intSl []int) (str []string) {
	for i := range intSl {
		number := intSl[i]
		text := strconv.Itoa(number)
		str = append(str, text)
	}
	return str
}

type Address struct {
	Country  string
	City     string
	Street   string
	Building string
}

func (a *Address) getAddress() string {
	country := countryList[randomInt(len(countryList))]
	index := indexList[randomInt(len(indexList))]
	city := cityList[randomInt(len(cityList))]
	srteetPrefix := srteetPrefixList[randomInt(len(srteetPrefixList))]
	srteet := srteetList[randomInt(len(srteetList))]
	return fmt.Sprintf("%s, %s, г. %s, %s %s, д. %d", country, index, city, srteetPrefix, srteet, randomInt(20))
}

func getMaleName() string {
	firstName := maleFirstNameList[randomInt(len(maleFirstNameList))]
	lastName := maleLastNameList[randomInt(len(maleLastNameList))]
	return fmt.Sprintf("%s %s", firstName, lastName)
}

func getFeMaleName() string {
	firstName := femaleFirstNameList[randomInt(len(femaleFirstNameList))]
	lastName := femaleLastNameList[randomInt(len(femaleLastNameList))]
	return fmt.Sprintf("%s %s", firstName, lastName)
}

func getEmail(name string) string {
	localPart := translit.EncodeToICAO(strings.ToLower(name))
	domainName := freeEmailList[randomInt(len(freeEmailList))]
	return fmt.Sprintf("%s@%s", strings.Join(strings.Fields(localPart), "."), domainName)
}

func getPhone(prefix string) string {
	randInt, _ := RandomInt(1, 10)
	str := strings.Join(IntToString(randInt), "")
	return fmt.Sprintf("%s (%s) %s-%s-%s", prefix, str[:3], str[3:6], str[6:8], str[8:10])
}

func getNumber() string {
	randInt, _ := RandomInt(1, 9)
	str := strings.Join(IntToString(randInt), "")
	return fmt.Sprintf("%s-%s/%s/%s", "IM", str[:3], str[3:6], str[6:9])
}

func getTitle() string {
	return titleList[randomInt(len(titleList))]
}

type Person struct {
	Name  string
	Email string
	Phone string
	Title string
}

func NewPerson(sex bool) *Person {
	var name string
	if sex {
		name = getMaleName()
	} else {
		name = getFeMaleName()
	}

	email := getEmail(name)
	title := getTitle()
	phone := getPhone("+7")
	return &Person{
		Name:  name,
		Email: email,
		Phone: phone,
		Title: title,
	}
}

type Company struct {
	Name     string
	FullName string
	INN      string
	CEO      string
	CFO      string
	Address  string
}

func NewCompany() *Company {
	name, fullName := getCompanyName()
	ceo := getMaleName()
	cfo := getFeMaleName()
	a := Address{}
	a.getAddress()

	return &Company{
		Name:     name,
		FullName: fullName,
		INN:      "12345678901",
		CEO:      ceo,
		CFO:      cfo,
		Address:  a.getAddress(),
	}
}

type Agreement struct {
	Name    string
	StartAt time.Time
}

func NewAgreement() *Agreement {
	start := time.Now()
	return &Agreement{
		Name:    getNumber(),
		StartAt: start.AddDate(0, -1*randomInt(10), 0),
	}
}

type Product struct {
	Name        string
	Description string
	SKU         string
	Price       float64
}

func getSKU(prefix string) string {
	randInt, _ := RandomInt(1, 4)
	str := strings.Join(IntToString(randInt), "")
	return fmt.Sprintf("%s-%s-%s", prefix, str[:2], str[2:4])
}

func ProductList() []Product {
	products := []Product{}
	price := randomInt(100) * 100
	for _, v := range productList {
		product := Product{
			Name:        v,
			Description: fmt.Sprintf("%s %s", nounList[randomInt(len(nounList))], v),
			SKU:         getSKU("AR"),
			Price:       float64(price),
		}
		products = append(products, product)
	}
	return products
}
