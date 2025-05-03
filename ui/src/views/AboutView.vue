<template>
  <div class="about-view">
    <div class="max-w-screen-md mx-auto">
      <h1 class="text-2xl font-bold mb-6">О проекте</h1>
      
      <div class="prose">
        <p class="mb-4">
          У сайта &laquo;Тысячники Южного Урала&raquo; две основные функции.
          Во-первых это удобный каталог всех вершин Южного Урала выше тысячи метров.
          Во-вторых это возможность каждому вести учет пройденных вершин.
          <p>
          Идея проекта позаимствована у неформального туристического клуба &laquo;Исилган&raquo;, который был создан в далеком 1982 году.
          Участники этого клуба ставили себе целью совершить восхождения на все вершины Южного Урала высотой более тысячи метров над уровнем моря.
          Историю клуба &laquo;Исилган&raquo; и причины его создания можно узнать <a href="http://www.isilgan.ru/2013/06/blog-post_5161.html">у его создателей</a>
          </p>
        </p>
        
        <h2 class="text-xl font-semibold mt-6 mb-4">Контактная информация</h2>
        <p class="mb-4">
          Новости проекта публикуются в группе <a href="https://vk.com/southural_1000">Вконтакте</a>
        </p>
        <p class="mb-4">
          По вопросам, касающимся работы сайта, можно писать на <a href="mailto:rush.ru@gmail.com">rush.ru@gmail.com</a>.
        </p>
        
        <h2 class="text-xl font-semibold mt-6 mb-4">Благодарности</h2>
        <p class="mb-4">
          Борису Прохорову — за идею и предоставленный каталог вершин.
        </p>
        <p class="mb-4">
          Евгению Безгодову — за администрирование проекта и фотографии вершин.
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
