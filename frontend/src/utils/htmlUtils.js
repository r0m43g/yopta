/**
 * Декодирование HTML-сущностей в обычный текст
 * Превращает закодированные символы типа &#34; в "
 * 
 * @param {string} text - Закодированный текст с HTML-сущностями
 * @returns {string} - Декодированный текст
 */
export function decodeHTMLEntities(text) {
  if (!text) return ''
  const textArea = document.createElement('textarea')
  textArea.innerHTML = text
  return textArea.value
}
