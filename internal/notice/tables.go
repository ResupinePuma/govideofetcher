package notice

import "errors"

var tableEng map[notice]string = map[notice]string{
	ErrSizeLimitReached.notice:   "🪨 Sorry, media is too large. I can't download it",
	ErrTimeout.notice:            "⌛ Sorry, timeout. External service takes a long time to respond or the video is too big. Try again later",
	ErrInvalidURL.notice:         "😤 URL incorrect. Fix and try again",
	ErrInvalidResponse.notice:    "🤔 Sorry, something happened and I can't download it now. Try again later",
	ErrNotFound.notice:           "👀 Sorry, I can't find media behind that link",
	ErrUnsupportedService.notice: "🛑 This service is prohibited due regional restrictons",
	ErrUnexpectedError.notice:    "💔 Oops, unexpected error. Try again later",

	NoticeMediaFound: "📣 I found %v media. Start sending",
	NoticeGotLink:    "Got link. 👀 at 📼",
	NoticeDone:       "Done! ✅",
}

var tableRus map[notice]string = map[notice]string{
	ErrSizeLimitReached.notice:   "🪨 Извините, медиафайл слишком большой. Я не могу его скачать",
	ErrInvalidURL.notice:         "😤 Неверный URL. Исправьте и попробуйте снова",
	ErrTimeout.notice:            "⌛ Извините, превышено время ожидания. Внешний сервис отвечает слишком долго или видео слишком большое. Попробуйте позже",
	ErrInvalidResponse.notice:    "🤔 Извините, что-то пошло не так, и я сейчас не могу скачать файл. Попробуйте ещё раз позже",
	ErrNotFound.notice:           "👀 Извините, не удалось найти медиафайл по этой ссылке",
	ErrUnsupportedService.notice: "🛑 Этот сервис недоступен из-за региональных ограничений",
	ErrUnexpectedError.notice:    "💔 Упс, произошла неожиданная ошибка. Попробуйте позже",

	NoticeMediaFound:       "📣 Я нашёл %v медиафайлов. Начинаю отправку",
	NoticeGotLink:          "Опа, ссылка. 👀 на 📼",
	NoticeDone:             "Всё! ✅",
	NoticeUsageGet:         "<Ссылка> <заголовок (опционально)> - скачать медиа по ссылке",
	NoticeUsageNotFound:    "команда не найдена",
	NoticeDownloadingMedia: "⏬ Скачивание медиа",
}

func translateError(err nerror, lang string) string {
	if lang == "" {
		return translateError(err, "en")
	}

	switch lang {
	case "ru":
		t, ok := tableRus[err.notice]
		if ok {
			return t
		} else {
			return tableRus[ErrUnexpectedError.notice]
		}
	default:
		t, ok := tableEng[err.notice]
		if ok {
			return t
		} else {
			return tableEng[ErrUnexpectedError.notice]
		}
	}
}

func TranslateError(err error, lang string) string {
	var parsed nerror
	if errors.As(err, &parsed) {
		return translateError(parsed, lang)
	}

	switch lang {
	case "ru":
		return tableRus[ErrUnexpectedError.notice]
	default:
		return tableEng[ErrUnexpectedError.notice]
	}
}

func TranslateNotice(text notice, lang string) string {
	switch lang {
	case "ru":
		t, ok := tableRus[text]
		if ok {
			return t
		}
	default:
		t, ok := tableEng[text]
		if ok {
			return t
		}
	}

	return text.String()
}
