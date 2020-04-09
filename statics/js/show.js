$(function(){
  var favs = [];
  var favNum = [];
  var spans = [];
  $("span").each(function(e,v){
      spans.push(v);
    });
    console.log(spans);
    for (let i=0; i < spans.length; i++){
      favNum.push(Number(spans[i].innerText));
    }
    favNum.sort(
      function(a,b){
        return (a < b ? 1 : -1);
      }
    );

    console.log(favNum[0]);
    console.log(spans[0].innerText);
    function GetNum1(){
      for (let i=0; i < spans.length; i++){
        if(favNum[0] == spans[i].innerText){
          return i;
        }
      }
    }
    function GetNum2(){
      for (let i=0; i < spans.length; i++){
        if(favNum[1] == spans[i].innerText){
          return i;
        }
      }
    }
    function GetNum3(){
      for (let i=0; i < spans.length; i++){
        if(favNum[2] == spans[i].innerText){
          return i;
        }
      }
    }
    function GetNum4(){
      for (let i=0; i < spans.length; i++){
        if(favNum[3] == spans[i].innerText){
          return i;
        }
      }
    }
    function GetNum5(){
      for (let i=0; i < spans.length; i++){
        if(favNum[4] == spans[i].innerText){
          return i;
        }
      }
    }
    console.log(favNum);

    if(spans[GetNum1()]){
      $('#' + `${spans[GetNum1()].id}`).parents('.picture').attr('style','grid-row:1/2;grid-column:1/4;');
      console.log(spans[GetNum1()].id);
      spans.splice(GetNum1(),1);
    }
    if(spans[GetNum2()]){
      $('#' + `${spans[GetNum2()].id}`).parents('.picture').attr('style','grid-row:2/3;grid-column:1/3;');
      console.log(spans[GetNum2()].id);
      spans.splice(GetNum2(),1);
    }
    if(spans[GetNum3()]){
      $('#' + `${spans[GetNum3()].id}`).parents('.picture').attr('style','grid-row:2/3;grid-column:3/4;');
      console.log(spans[GetNum3()].id);
      spans.splice(GetNum3(),1);
    }
    if(spans[GetNum4()]){
      $('#' + `${spans[GetNum4()].id}`).parents('.picture').attr('style','grid-row:3/4;grid-column:1/2;');
      console.log(spans[GetNum4()].id);
      spans.splice(GetNum4(),1);
    }
    if(spans[GetNum5()]){
      $('#' + `${spans[GetNum5()].id}`).parents('.picture').attr('style','grid-row:3/4;grid-column:2/4;');
      console.log(spans[GetNum5()].id);
      spans.splice(GetNum5(),1);
    }

  $('div[id^="favorite"]').on("click", function(){
    var path = $(this).attr('id').replace('favorite-', '');
    $.ajax({
      type: 'POST',
      url: '/favorite',
      data: {
        "path": path
      }
    }).then(
      data => {
        id = '#favorite-count-'+path;
        $(id).html(String(data));
      },
      error => alert("Error")
    )
    location.reload();
  });
});
