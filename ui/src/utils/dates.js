/**
 * Formats a date object according to specified rules:
 * 1. If all fields are non-zero: "2 Января 2006"
 * 2. If only month and year are non-zero: "Январь 2006"
 * 3. If only year is non-zero: "2006"
 * 4. If all fields are zero: empty string
 */
export function formatRussianDate(date) {
  if (!date) return '';
  
  const { Year, Month, Day } = date;
  
  // If all fields are zero, return empty string
  if (!Year && !Month && !Day) return '';
  
  // If only year is non-zero
  if (Year && !Month && !Day) {
    return `${Year}`;
  }
  
  // Russian month names
  const monthNames = [
    'Января', 'Февраля', 'Марта', 'Апреля', 'Мая', 'Июня',
    'Июля', 'Августа', 'Сентября', 'Октября', 'Ноября', 'Декабря'
  ];
  
  const monthNamesNominative = [
    'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
    'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
  ];
  
  // If month and year are non-zero, but day is zero
  if (Year && Month && !Day) {
    return `${monthNamesNominative[Month - 1]} ${Year}`;
  }
  
  // If all fields are non-zero
  if (Year && Month && Day) {
    return `${Day} ${monthNames[Month - 1]} ${Year}`;
  }
  
  return '';
} 