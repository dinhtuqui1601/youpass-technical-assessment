export interface PracticeOption {
  key: string
  text: string
}

export interface PracticeQuestion {
  id: number
  text: string
  options: PracticeOption[]
}

// Demo questions. The backend has no exam module yet, so this data is clearly
// illustrative and local only (never persisted to the server).
export const demoQuestions: PracticeQuestion[] = [
  {
    id: 1,
    text: 'Thẻ HTML nào dùng để tạo một liên kết (hyperlink)?',
    options: [
      { key: 'A', text: '<a>' },
      { key: 'B', text: '<link>' },
      { key: 'C', text: '<href>' },
      { key: 'D', text: '<url>' },
    ],
  },
  {
    id: 2,
    text: 'Thuộc tính CSS nào thay đổi màu chữ?',
    options: [
      { key: 'A', text: 'font-style' },
      { key: 'B', text: 'color' },
      { key: 'C', text: 'text-color' },
      { key: 'D', text: 'font-color' },
    ],
  },
  {
    id: 3,
    text: 'Trong JavaScript, từ khóa nào khai báo biến bất biến (không gán lại)?',
    options: [
      { key: 'A', text: 'var' },
      { key: 'B', text: 'let' },
      { key: 'C', text: 'const' },
      { key: 'D', text: 'static' },
    ],
  },
  {
    id: 4,
    text: 'HTTP status code nào thể hiện "Not Found"?',
    options: [
      { key: 'A', text: '200' },
      { key: 'B', text: '301' },
      { key: 'C', text: '404' },
      { key: 'D', text: '500' },
    ],
  },
  {
    id: 5,
    text: 'Cấu trúc dữ liệu nào hoạt động theo nguyên tắc LIFO?',
    options: [
      { key: 'A', text: 'Queue' },
      { key: 'B', text: 'Stack' },
      { key: 'C', text: 'Linked List' },
      { key: 'D', text: 'Hash Map' },
    ],
  },
  {
    id: 6,
    text: 'Trong Git, lệnh nào dùng để lưu thay đổi vào lịch sử?',
    options: [
      { key: 'A', text: 'git add' },
      { key: 'B', text: 'git commit' },
      { key: 'C', text: 'git push' },
      { key: 'D', text: 'git status' },
    ],
  },
]
