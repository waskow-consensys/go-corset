(column bin:ARGUMENT_1_HI :u128)
(column bin:XXX_BYTE_LO :u8)
(column bin:INST :u8)
(column bin:ACC_6 :u128)
(column bin:IS_BYTE :u1)
(column bin:BIT_1 :u1)
(column bin:COUNTER :u8)
(column bin:ARGUMENT_2_HI :u128)
(column bin:ARGUMENT_2_LO :u128)
(column bin:IS_AND :u1)
(column bin:BYTE_2 :u8)
(column bin:ARGUMENT_1_LO :u128)
(column bin:ACC_3 :u128)
(column bin:IS_NOT :u1)
(column bin:BITS :u1)
(column bin:BIT_B_4 :u1)
(column bin:ACC_5 :u128)
(column bin:CT_MAX :u8)
(column bin:BYTE_3 :u8)
(column bin:BYTE_5 :u8)
(column bin:RESULT_LO :u128)
(column bin:SMALL :u1)
(column bin:IS_SIGNEXTEND :u1)
(column bin:LOW_4 :u8)
(column bin:ACC_2 :u128)
(column bin:IS_XOR :u1)
(column bin:BYTE_6 :u8)
(column bin:RESULT_HI :u128)
(column bin:NEG :u1)
(column bin:STAMP :u32)
(column bin:BYTE_1 :u8)
(column bin:XXX_BYTE_HI :u8)
(column bin:ACC_1 :u128)
(column bin:PIVOT :u8)
(column bin:BYTE_4 :u8)
(column bin:ACC_4 :u128)
(column bin:IS_OR :u1)
(vanish bin:byte_decompositions (begin (if bin:COUNTER (- bin:ACC_1 bin:BYTE_1) (- bin:ACC_1 (+ (* 256 (shift bin:ACC_1 -1)) bin:BYTE_1))) (if bin:COUNTER (- bin:ACC_2 bin:BYTE_2) (- bin:ACC_2 (+ (* 256 (shift bin:ACC_2 -1)) bin:BYTE_2))) (if bin:COUNTER (- bin:ACC_3 bin:BYTE_3) (- bin:ACC_3 (+ (* 256 (shift bin:ACC_3 -1)) bin:BYTE_3))) (if bin:COUNTER (- bin:ACC_4 bin:BYTE_4) (- bin:ACC_4 (+ (* 256 (shift bin:ACC_4 -1)) bin:BYTE_4))) (if bin:COUNTER (- bin:ACC_5 bin:BYTE_5) (- bin:ACC_5 (+ (* 256 (shift bin:ACC_5 -1)) bin:BYTE_5))) (if bin:COUNTER (- bin:ACC_6 bin:BYTE_6) (- bin:ACC_6 (+ (* 256 (shift bin:ACC_6 -1)) bin:BYTE_6)))))
(vanish bin:bits-and-related (ifnot (+ bin:IS_BYTE bin:IS_SIGNEXTEND) (if (- bin:COUNTER 15) (begin (- bin:PIVOT (+ (* 128 (shift bin:BITS -15)) (* 64 (shift bin:BITS -14)) (* 32 (shift bin:BITS -13)) (* 16 (shift bin:BITS -12)) (* 8 (shift bin:BITS -11)) (* 4 (shift bin:BITS -10)) (* 2 (shift bin:BITS -9)) (shift bin:BITS -8))) (- bin:BYTE_2 (+ (* 128 (shift bin:BITS -7)) (* 64 (shift bin:BITS -6)) (* 32 (shift bin:BITS -5)) (* 16 (shift bin:BITS -4)) (* 8 (shift bin:BITS -3)) (* 4 (shift bin:BITS -2)) (* 2 (shift bin:BITS -1)) bin:BITS)) (- bin:LOW_4 (+ (* 8 (shift bin:BITS -3)) (* 4 (shift bin:BITS -2)) (* 2 (shift bin:BITS -1)) bin:BITS)) (- bin:BIT_B_4 (shift bin:BITS -4)) (- bin:NEG (shift bin:BITS -15))))))
(vanish bin:pivot (ifnot bin:CT_MAX (begin (if (- bin:IS_BYTE 1) (if bin:LOW_4 (if bin:COUNTER (if bin:BIT_B_4 (- bin:PIVOT bin:BYTE_3) (- bin:PIVOT bin:BYTE_4))) (if (+ (shift bin:BIT_1 -1) (- 1 bin:BIT_1)) (if bin:BIT_B_4 (- bin:PIVOT bin:BYTE_3) (- bin:PIVOT bin:BYTE_4))))) (if (- bin:IS_SIGNEXTEND 1) (if (- bin:LOW_4 15) (if bin:COUNTER (if bin:BIT_B_4 (- bin:PIVOT bin:BYTE_4) (- bin:PIVOT bin:BYTE_3))) (if (+ (shift bin:BIT_1 -1) (- 1 bin:BIT_1)) (if bin:BIT_B_4 (- bin:PIVOT bin:BYTE_4) (- bin:PIVOT bin:BYTE_3))))))))
(vanish bin:counter-constancies (begin (ifnot bin:COUNTER (- bin:ARGUMENT_1_HI (shift bin:ARGUMENT_1_HI -1))) (ifnot bin:COUNTER (- bin:ARGUMENT_1_LO (shift bin:ARGUMENT_1_LO -1))) (ifnot bin:COUNTER (- bin:ARGUMENT_2_HI (shift bin:ARGUMENT_2_HI -1))) (ifnot bin:COUNTER (- bin:ARGUMENT_2_LO (shift bin:ARGUMENT_2_LO -1))) (ifnot bin:COUNTER (- bin:RESULT_HI (shift bin:RESULT_HI -1))) (ifnot bin:COUNTER (- bin:RESULT_LO (shift bin:RESULT_LO -1))) (ifnot bin:COUNTER (- bin:INST (shift bin:INST -1))) (ifnot bin:COUNTER (- bin:CT_MAX (shift bin:CT_MAX -1))) (ifnot bin:COUNTER (- bin:PIVOT (shift bin:PIVOT -1))) (ifnot bin:COUNTER (- bin:BIT_B_4 (shift bin:BIT_B_4 -1))) (ifnot bin:COUNTER (- bin:LOW_4 (shift bin:LOW_4 -1))) (ifnot bin:COUNTER (- bin:NEG (shift bin:NEG -1))) (ifnot bin:COUNTER (- bin:SMALL (shift bin:SMALL -1)))))
(vanish bin:bit_1 (ifnot bin:CT_MAX (begin (if (- bin:IS_BYTE 1) (begin (if bin:LOW_4 (- bin:BIT_1 1) (if (- bin:COUNTER 0) bin:BIT_1 (if (- bin:COUNTER bin:LOW_4) (- bin:BIT_1 (+ (shift bin:BIT_1 -1) 1)) (- bin:BIT_1 (shift bin:BIT_1 -1))))))) (if (- bin:IS_SIGNEXTEND 1) (begin (if (- 15 bin:LOW_4) (- bin:BIT_1 1) (if (- bin:COUNTER 0) bin:BIT_1 (if (- bin:COUNTER (- 15 bin:LOW_4)) (- bin:BIT_1 (+ (shift bin:BIT_1 -1) 1)) (- bin:BIT_1 (shift bin:BIT_1 -1))))))))))
(vanish bin:is-signextend-result (ifnot bin:IS_SIGNEXTEND (if bin:CT_MAX (begin (- bin:RESULT_HI bin:ARGUMENT_2_HI) (- bin:RESULT_LO bin:ARGUMENT_2_LO)) (if bin:SMALL (begin (- bin:RESULT_HI bin:ARGUMENT_2_HI) (- bin:RESULT_LO bin:ARGUMENT_2_LO)) (begin (if bin:BIT_B_4 (begin (- bin:BYTE_5 (* bin:NEG 255)) (if bin:BIT_1 (- bin:BYTE_6 (* bin:NEG 255)) (- bin:BYTE_6 bin:BYTE_4))) (begin (if bin:BIT_1 (- bin:BYTE_5 (* bin:NEG 255)) (- bin:BYTE_5 bin:BYTE_3)) (- bin:RESULT_LO bin:ARGUMENT_2_LO))))))))
(vanish bin:small (ifnot (+ bin:IS_BYTE bin:IS_SIGNEXTEND) (if (- bin:COUNTER 15) (if bin:ARGUMENT_1_HI (if (- bin:ARGUMENT_1_LO (+ (* 16 (shift bin:BITS -4)) (* 8 (shift bin:BITS -3)) (* 4 (shift bin:BITS -2)) (* 2 (shift bin:BITS -1)) bin:BITS)) (- bin:SMALL 1) bin:SMALL)))))
(vanish bin:inst-to-flag (- bin:INST (+ (* bin:IS_AND 22) (* bin:IS_OR 23) (* bin:IS_XOR 24) (* bin:IS_NOT 25) (* bin:IS_BYTE 26) (* bin:IS_SIGNEXTEND 11))))
(vanish bin:target-constraints (if (- bin:COUNTER bin:CT_MAX) (begin (- bin:ACC_1 bin:ARGUMENT_1_HI) (- bin:ACC_2 bin:ARGUMENT_1_LO) (- bin:ACC_3 bin:ARGUMENT_2_HI) (- bin:ACC_4 bin:ARGUMENT_2_LO) (- bin:ACC_5 bin:RESULT_HI) (- bin:ACC_6 bin:RESULT_LO))))
(vanish bin:countereset (ifnot bin:STAMP (if (- bin:COUNTER bin:CT_MAX) (- (shift bin:STAMP 1) (+ bin:STAMP 1)) (- (shift bin:COUNTER 1) (+ bin:COUNTER 1)))))
(vanish bin:stamp-increments (* (- (shift bin:STAMP 1) (+ bin:STAMP 0)) (- (shift bin:STAMP 1) (+ bin:STAMP 1))))
(vanish bin:isbyte-ctmax (if (- (+ bin:IS_BYTE bin:IS_SIGNEXTEND) 1) (if bin:ARGUMENT_1_HI (- bin:CT_MAX 15) bin:CT_MAX)))
(vanish bin:no-bin-no-flag (if bin:STAMP (+ bin:IS_AND bin:IS_OR bin:IS_XOR bin:IS_NOT bin:IS_BYTE bin:IS_SIGNEXTEND) (- (+ bin:IS_AND bin:IS_OR bin:IS_XOR bin:IS_NOT bin:IS_BYTE bin:IS_SIGNEXTEND) 1)))
(vanish bin:is-byte-result (ifnot bin:IS_BYTE (if bin:CT_MAX (begin bin:RESULT_HI bin:RESULT_LO) (begin bin:RESULT_HI (- bin:RESULT_LO (* bin:SMALL bin:PIVOT))))))
(vanish bin:result-via-lookup (ifnot (+ bin:IS_AND bin:IS_OR bin:IS_XOR bin:IS_NOT) (begin (- bin:BYTE_5 bin:XXX_BYTE_HI) (- bin:BYTE_6 bin:XXX_BYTE_LO))))
(vanish bin:isnot-ctmax (if (- bin:IS_NOT 1) (- bin:CT_MAX 15)))
(vanish bin:ct-small (- 1 (~ (- bin:COUNTER 16))))
(vanish bin:new-stamp-reset-ct (ifnot (- (shift bin:STAMP 1) bin:STAMP) (shift bin:COUNTER 1)))
(vanish:last bin:last-row (- bin:COUNTER bin:CT_MAX))
(vanish:first bin:first-row bin:STAMP)