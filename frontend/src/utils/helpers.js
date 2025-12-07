// src/models/TrainModels.js
export function timeToDecimal(timeStr) {
  const start = new Date('2025-01-01T00:00:00Z'); // Начало эпохи
  if (!timeStr) return null;

  if (timeStr instanceof Date) {
    let diff = (timeStr - start) / 1000 / 60; // разница в минутах
    return Math.floor(diff);
  }

  if (typeof timeStr === 'string') {
    if (timeStr.includes('T')) {
      const date = new Date(timeStr);
      if (!isNaN(date.getTime())) {
        let diff = (date - start) / 1000 / 60; // разница в минутах
        return Math.floor(diff);
      }
    }

  }

  return null;
}

