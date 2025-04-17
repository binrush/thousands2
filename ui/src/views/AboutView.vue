<template>
  <div class="about-view">
    <div class="max-w-screen-md mx-auto">
      <h1 class="text-2xl font-bold mb-6">О проекте</h1>
      
      <div class="prose">
        <p class="mb-4">
          Проект "1000+" создан для альпинистов и любителей гор. Здесь вы можете:
        </p>
        
        <ul class="list-disc pl-6 mb-4">
          <li>Изучать информацию о вершинах Южного Урала</li>
          <li>Отслеживать свои восхождения</li>
          <li>Делиться фотографиями и впечатлениями</li>
          <li>Соревноваться с другими альпинистами</li>
        </ul>

        <h2 class="text-xl font-semibold mt-6 mb-4">Как это работает?</h2>
        <p class="mb-4">
          Каждая вершина в базе данных имеет свою страницу с подробной информацией:
          высота, расположение, описание маршрутов. После восхождения вы можете
          зарегистрировать его, добавить фотографии и описание.
        </p>

        <h2 class="text-xl font-semibold mt-6 mb-4">Рейтинг</h2>
        <p class="mb-4">
          Система автоматически подсчитывает ваш рейтинг на основе количества
          восхождений и общей высоты покоренных вершин. Чем больше вершин вы
          покорили, тем выше ваше место в рейтинге.
        </p>

        <h2 class="text-xl font-semibold mt-6 mb-4">Контакты</h2>
        <p>
          Если у вас есть вопросы или предложения, пожалуйста, свяжитесь с нами:
          <a href="mailto:info@1000plus.ru" class="text-blue-600 hover:underline">info@1000plus.ru</a>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
// No additional logic needed for this view
</script>

<style scoped>
.about-view {
  padding: 20px;
}

.prose {
  max-width: 65ch;
  color: #374151;
}

.prose a {
  color: #2563eb;
}

.prose a:hover {
  text-decoration: underline;
}
</style>
