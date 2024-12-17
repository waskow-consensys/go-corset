;; [TODO] (defunalias debug-assert debug)

(defpurefun (if-zero cond then) (if (vanishes! cond) then))
;; [TODO] (defpurefun (if-zero cond then else) (if (vanishes! cond) then else))
(defpurefun (if-zero-else cond then else) (if (vanishes! cond) then else))

(defpurefun (if-not-zero cond then) (if (force-bool cond) then))
;; [TODO] (defpurefun (if-not-zero cond then else) (if (force-bool cond) then else))

(defpurefun ((force-bool :@bool :force) x) x)
(defpurefun ((is-binary :@loob :force) e0) (* e0 (- 1 e0)))

(defpurefun ((force-bin :binary :force) x) x)

;;
;; Boolean functions
;;
;; !-suffix denotes loobean algebra (i.e. 0 == true)
;; ~-prefix denotes normalized-functions (i.e. output is 0/1)
(defpurefun (and a b) (* a b))
;; [TODO] (defpurefun ((~and :binary@bool) a b) (~ (and a b)))
(defpurefun ((or! :@loob) a b) (* a b))
;; [TODO] (defpurefun ((~or! :binary@loob) a b) (~ (or! a b)))

;; [TODO] (defpurefun ((not :binary@bool :force) (x :binary)) (- 1 x))

(defpurefun ((eq! :@loob) x y) (- x y))
;; [TODO] (defpurefun ((neq! :binary@loob :force) x y) (not (~ (eq! x y))))
;; [TODO] (defunalias = eq!)

;; [TODO] (defpurefun ((eq :binary@bool :force) (x :binary) (y :binary)) (^ (- x y) 2))
(defpurefun ((eq :binary@bool :force) x y) (- 1 (~ (eq! x y))))
(defpurefun ((neq :binary@bool :force) x y) (eq! x y))

;; Variadic variations on and/or
;; [TODO] (defunalias any! *)
;; [TODO] (defunalias all *)

;; Boolean functions
(defpurefun ((is-not-zero :binary@bool) x) (~ x))
(defpurefun ((is-not-zero! :binary@loob :force) x) (- 1 (is-not-zero x)))
(defpurefun ((is-zero :binary@bool :force) x) (- 1 (~ x)))

;; Chronological functions
(defpurefun (next X) (shift X 1))
(defpurefun (prev X) (shift X -1))

;; Ensure that e0 has (resp. will) increase (resp. decrease) of offset
;; w.r.t. the previous (resp. next) row.
(defpurefun (did-inc! e0 offset) (eq! e0 (+ (prev e0) offset)))
(defpurefun (did-dec! e0 offset) (eq!  e0 (- (prev e0) offset)))
(defpurefun (will-inc! e0 offset) (will-eq! e0 (+ e0 offset)))
(defpurefun (will-dec! e0 offset) (eq! (next e0) (- e0 offset)))

(defpurefun (did-inc e0 offset) (eq e0 (+ (prev e0) offset)))
(defpurefun (did-dec e0 offset) (eq  e0 (- (prev e0) offset)))
(defpurefun (will-inc e0 offset) (will-eq e0 (+ e0 offset)))
(defpurefun (will-dec e0 offset) (eq (next e0) (- e0 offset)))

;; Ensure that e0 remained (resp. will be) constant
;; with regards to the previous (resp. next) row.
(defpurefun (remained-constant! e0) (eq! e0 (prev e0)))
(defpurefun (will-remain-constant! e0) (will-eq! e0 e0))

(defpurefun (remained-constant e0) (eq e0 (prev e0)))
(defpurefun (will-remain-constant e0) (will-eq e0 e0))

;; Ensure (in loobean logic) that e0 has changed (resp. will change) its value
;; with regards to the previous (resp. next) row.
;; [TODO] (defpurefun (did-change! e0) (neq! e0 (prev e0)))
;; [TODO] (defpurefun (will-change! e0) (neq! e0 (next e0)))

(defpurefun (did-change e0) (neq e0 (prev e0)))
(defpurefun (will-change e0) (neq e0 (next e0)))

;; Ensure (in loobean logic) that e0 was (resp. will be) equal to e1 in the
;; previous (resp. next) row.
(defpurefun (was-eq! e0 e1) (eq! (prev e0) e1))
(defpurefun (will-eq! e0 e1) (eq! (next e0) e1))

(defpurefun (was-eq e0 e1) (eq (prev e0) e1))
(defpurefun (will-eq e0 e1) (eq (next e0) e1))


;; Helpers
(defpurefun ((vanishes! :@loob :force) e0) e0)
(defpurefun (if-eq x val then) (if (eq! x val) then))
(defpurefun (if-eq-else x val then else) (if (eq! x val) then else))

;; counter constancy constraint
(defpurefun ((counter-constancy :@loob) ct X)
  (if-not-zero ct
               (remained-constant! X)))

;; perspective constancy constraint
(defpurefun ((perspective-constancy :@loob) PERSPECTIVE_SELECTOR X)
            (if-not-zero (* PERSPECTIVE_SELECTOR (prev PERSPECTIVE_SELECTOR))
                         (remained-constant! X)))

;; base-X decomposition constraints
(defpurefun (base-X-decomposition ct base acc digits)
  (if-zero-else ct
           (eq! acc digits)
           (eq! acc (+ (* base (prev acc)) digits))))

;; byte decomposition constraint
(defpurefun (byte-decomposition ct acc bytes) (base-X-decomposition ct 256 acc bytes))

;; bit decomposition constraint
(defpurefun (bit-decomposition ct acc bits) (base-X-decomposition ct 2 acc bits))

;; plateau constraints
;; [TODO] (defpurefun (plateau-constraint CT (X :binary) C)
;;             (begin (debug-assert (stamp-constancy CT C))
;;                    (if-zero C
;;                             (eq! X 1)
;;                             (if (eq! CT 0)
;;                                 (vanishes! X)
;;                               (if (eq!  CT C)
;;                                   (did-inc! X 1)
;;                                 (remained-constant! X))))))

;; stamp constancy imposes that the column C may only
;; change at rows where the STAMP column changes.
(defpurefun (stamp-constancy STAMP C)
            (if (will-remain-constant! STAMP)
                (will-remain-constant! C)))

;; =============================================================================
;; Add
;; =============================================================================
(defpurefun (if-not-eq X Y then)
    (if (eq! X Y)
        ;; True branch
        (vanishes! 0)
        ;; False branch
        then))