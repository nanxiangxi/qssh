(function ($) {
    'use strict';

    setTimeout(function() {
        gsap.registerPlugin(ScrollTrigger);

        // animation1
        const tm_gsap_split_text = document.querySelector(".tm-split-text");
        if(tm_gsap_split_text) {
            setTimeout(function () {
                var st = $('document').find(".tm-split-text");
                if(st.length == 0) return;
                gsap.registerPlugin(SplitText);
                st.each(function(index, el) {
                    el.split = new SplitText(el, {
                        type: "lines,words,chars",
                        linesClass: "split-line"
                    });
                    gsap.set(el, { perspective: 400 });

                    if( $(el).hasClass('split-in-fade') ){
                        $(el).addClass('active');
                        gsap.set(el.split.chars, {
                            opacity: 0,
                            ease: "Back.easeOut",
                        });
                    }
                    if( $(el).hasClass('split-in-right') ){
                        gsap.set(el.split.chars, {
                            opacity: 0,
                            x: "50",
                            ease: "Back.easeOut",
                        });
                    }
                    if( $(el).hasClass('split-in-left') ){
                        gsap.set(el.split.chars, {
                            opacity: 0,
                            x: "-50",
                            ease: "circ.out",
                        });
                    }
                    if( $(el).hasClass('split-in-up') ){
                        gsap.set(el.split.chars, {
                            opacity: 0,
                            y: "80",
                            ease: "circ.out",
                        });
                    }
                    if( $(el).hasClass('split-in-down') ){
                        gsap.set(el.split.chars, {
                            opacity: 0,
                            y: "-80",
                            ease: "circ.out",
                        });
                    }
                    if( $(el).hasClass('split-in-rotate') ){
                        gsap.set(el.split.chars, {
                            opacity: 0,
                            rotateX: "50deg",
                            ease: "circ.out",
                        });
                    }
                    if( $(el).hasClass('split-in-scale') ){
                        gsap.set(el.split.chars, {
                            opacity: 0,
                            scale: "0.5",
                            ease: "circ.out",
                        });
                    }
                    el.anim = gsap.to(el.split.chars, {
                        scrollTrigger: {
                            trigger: el,
                            toggleActions: "restart pause resume reverse",
                            start: "top 90%",
                        },
                        x: "0",
                        y: "0",
                        rotateX: "0",
                        scale: 1,
                        opacity: 1,
                        duration: 0.8,
                        stagger: 0.02,
                    });
                });
            }, 200);
        }




        // animation1
        const tm_gsap_move_animation1 = document.querySelector(".tm-gsap-move-animation1");
        if(tm_gsap_move_animation1) {
            var tm_gsap_animation1 = gsap.timeline({
                scrollTrigger: {
                animation: tm_gsap_animation1,
                trigger: tm_gsap_move_animation1,
                start: "top 95%",
                end: "top -50%",
                scrub: 4,
                toggleActions: "play reverse play reverse",
                markers: false
                }
            });
            tm_gsap_animation1.from(tm_gsap_move_animation1, { xPercent: 50, });
        }




        // animation2
        const tm_gsap_move_animation2 = document.querySelector(".tm-gsap-move-animation2");
        if(tm_gsap_move_animation2) {
            var tm_gsap_animation2 = gsap.timeline({
                scrollTrigger: {
                    animation: tm_gsap_animation2,
                    trigger: tm_gsap_move_animation2,
                    start: "top 150%",
                    end: "top -50%",
                    scrub: 3,
                    toggleActions: "play reverse play reverse",
                    markers: false
                }
            });
            tm_gsap_animation2.from(tm_gsap_move_animation2, { xPercent: 50, yPercent: -10, scale: .3 }, "<=.5");
        }



        // animation3
        const tm_gsap_move_animation3 = document.querySelector(".tm-gsap-move-animation3");
        if(tm_gsap_move_animation3) {
            var tm_gsap_animation3 = gsap.timeline({
                scrollTrigger: {
                animation: tm_gsap_animation3,
                trigger: tm_gsap_move_animation3,
                start: "top 150%",
                end: "top -50%",
                scrub: 3,
                toggleActions: "play reverse play reverse",
                markers: false
                }
            });
            tm_gsap_animation3.from(tm_gsap_move_animation3, { xPercent: 70, yPercent: -50, scale: .3, rotate: -20, opacity: 0.3 }, "<=.5");
        }

    }, 2300);




    gsap.utils.toArray('.tm-gsap-animate-left').forEach((el, index) => {
        let tlcta = gsap.timeline({
            scrollTrigger: {
                trigger: el,
                scrub: 2,
                start: "top 90%",
                end: "top 70%",
                toggleActions: "play none none reverse",
                markers: false
            }
        })

        tlcta
        .set(el, {transformOrigin: 'center center'})
        .from(el, { opacity: 1,  x: "-=150"}, {opacity: 1, x: 0, duration: 1, immediateRender: false})
    });
    gsap.utils.toArray('.tm-gsap-animate-right').forEach((el, index) => {
        let tlcta = gsap.timeline({
            scrollTrigger: {
                trigger: el,
                scrub: 2,
                start: "top 90%",
                end: "top 70%",
                toggleActions: "play none none reverse",
                markers: false
            }
        })

        tlcta
        .set(el, {transformOrigin: 'center center'})
        .from(el, { opacity: 1,  x: "+=150"}, {opacity: 1, x: 0, duration: 1, immediateRender: false})
    });
    gsap.utils.toArray('.tm-gsap-animate-top').forEach((el, index) => {
        let tlcta = gsap.timeline({
            scrollTrigger: {
                trigger: el,
                scrub: 2,
                start: "top 90%",
                end: "top 70%",
                toggleActions: "play none none reverse",
                markers: false
            }
        })

        tlcta
        .set(el, {transformOrigin: 'center center'})
        .from(el, { opacity: 1,  y: "+=150"}, {opacity: 1, y: 0, duration: 1, immediateRender: false})
    });
    gsap.utils.toArray('.tm-gsap-animate-bottom').forEach((el, index) => {
        let tlcta = gsap.timeline({
            scrollTrigger: {
                trigger: el,
                scrub: 2,
                start: "top 90%",
                end: "top 70%",
                toggleActions: "play none none reverse",
                markers: false
            }
        })

        tlcta
        .set(el, {transformOrigin: 'center center'})
        .from(el, { opacity: 1,  y: "-=150"}, {opacity: 1, y: 0, duration: 1, immediateRender: false})
    });


    gsap.utils.toArray('.tm-gsap-animate-circle').forEach((el, index) => {
        let arspin = gsap.timeline({
            scrollTrigger: {
                trigger: el,
                scrub: 1,
                start: "top 100%",
                end: "top -50%",
                toggleActions: "play none none reverse",
                markers: false
            }
        })

        arspin
        .set(el, {transformOrigin: 'center center'})
        .fromTo(el, { rotate: 0}, { rotate: 180, duration: 2, immediateRender: false})
    });

    gsap.utils.toArray('.tm-gsap-animate-horizontal').forEach((el, index) => {
        let arspin = gsap.timeline({
            scrollTrigger: {
                trigger: el,
                scrub: 1,
                start: "top 100%",
                end: "top -50%",
                toggleActions: "play none none reverse",
                markers: false
            }
        })

        arspin
        .set(el, {transformOrigin: 'center center'})
        .fromTo(el, { xPercent: -20}, { xPercent: 50, duration: 2, immediateRender: false})
    });


    gsap.utils.toArray(".tm-gsap-img-parallax").forEach(function(container) {
      let image = container.querySelector("img");

      let tl = gsap.timeline({
          scrollTrigger: {
              trigger: container,
              scrub: .5,
          },
      });
      tl.from(image, {
          yPercent: -30,
          ease: "none",
      }).to(image, {
          yPercent: 30,
          ease: "none",
      });
    });

})(jQuery);