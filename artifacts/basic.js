
async function get_likes(tag) {
fetch('https://nhentai.net/api/gallery/' + tag, {
      method: 'GET',
      withCredentials: true,
      crossorigin: true,
      mode: 'no-cors',
    })
      .then((res) => res.json())
      .then((data) => {
        console.log(data);
      })
      .catch((error) => {
        console.error(error);
      });




    //div = document.getElementById(tag)
    //div.innerHTML = '(♡' + (await (await fetch('https://nhentai.net/api/gallery/' + tag)).json()).num_favorites + ')'
}
