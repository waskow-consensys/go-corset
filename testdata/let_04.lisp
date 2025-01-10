(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (X :@loob) Y Z)
(defconstraint c1 ()
  (let ((XeqY (- X Y)))
    (if X
        ;; if X==0 then Y == Z
        (begin
         (vanishes! Y)
         (vanishes! Z))
        ;; else X == Y && (Y == 0 || Z == 0)
        (begin
         (vanishes! XeqY)
         (vanishes! (* Y Z))))))
  ;; Z is always 0!
(defproperty a1 Z)