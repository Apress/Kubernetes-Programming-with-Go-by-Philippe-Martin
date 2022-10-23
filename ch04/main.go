package main

import (
	"fmt"

	"gopkg.in/inf.v0"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func main() {
	// # Pointers
	// ## Getting the reference of a value
	spec := appsv1.DeploymentSpec{
		Replicas: pointer.Int32(3),
	}

	// ## Dereferencing a pointer
	replicas := pointer.Int32Deref(spec.Replicas, 1)
	_ = replicas

	// ## Comparing two referenced values
	spec1 := appsv1.DeploymentSpec{
		Replicas: pointer.Int32(3),
	}

	spec2 := appsv1.DeploymentSpec{
		Replicas: pointer.Int32(1),
	}

	eq := pointer.Int32Equal(
		spec1.Replicas,
		spec2.Replicas,
	)
	_ = eq

	// # Quantities
	// ## Parsing a string as Quantity
	q1 := resource.MustParse("1Mi")
	q1, err := resource.ParseQuantity("1Mi")
	_ = q1
	_ = err

	// ## Using an inf.Dec as Quantity
	newDec := inf.NewDec(4, 3)
	fmt.Printf("newDec: %s\n", newDec) // 0.004
	q2 := resource.NewDecimalQuantity(*newDec, resource.DecimalExponent)
	fmt.Printf("q2: %s\n", q2) // 4e-3

	// ## Using a scaled integer as Quantity
	q3 := resource.NewScaledQuantity(4, 3)
	fmt.Printf("q3: %s\n", q3) // 4k

	q3.SetScaled(5, 6)
	fmt.Printf("q3: %s\n", q3) // 5M

	fmt.Printf("q3 scaled to 3: %d\n", q3.ScaledValue(3)) // 5.000
	fmt.Printf("q3 scaled to 0: %d\n", q3.ScaledValue(0)) // 5.000.000

	q4 := resource.NewQuantity(4000, resource.DecimalSI)
	fmt.Printf("q4: %s\n", q4) // 4k

	q5 := resource.NewQuantity(1024, resource.BinarySI)
	fmt.Printf("q5: %s\n", q5) // 1Ki

	q6 := resource.NewQuantity(4000, resource.DecimalExponent)
	fmt.Printf("q6: %s\n", q6) // 4e3

	q7 := resource.NewMilliQuantity(5, resource.DecimalSI)
	fmt.Printf("q7: %s\n", q7) // 5m

	q8 := resource.NewMilliQuantity(5, resource.DecimalExponent)
	fmt.Printf("q8: %s\n", q8) // 5e-3

	q8.SetMilli(6)
	fmt.Printf("q8: %s\n", q8) // 6e-3

	fmt.Printf("milli value of q8: %d\n", q8.MilliValue()) // 6

	// ## Operations on Quantities
	q9 := resource.MustParse("4M")
	q10 := resource.MustParse("3M")

	q9.Add(q10)
	fmt.Printf("4M + 3M: %s\n", q9.String()) // 7M

	q9.Sub(q10)
	fmt.Printf("7M - 3M: %s\n", q9.String()) // 4M

	cmp := q9.Cmp(q10)
	fmt.Printf("4M >? 3M: %d\n", cmp) // 1

	cmp = q9.CmpInt64(4_000_000)
	fmt.Printf("4M >? 4.000.000: %d\n", cmp) // 0

	q9.Neg()
	fmt.Printf("negative of 4M: %s\n", q9.String()) // -4M

	eq = q9.Equal(q10)
	fmt.Printf("4M ==? 3M: %v\n", eq) // false

	// # IntOrString
	ios1 := intstr.FromInt(10)
	fmt.Printf("ios1: %s\n", ios1.String()) // 10

	ios2 := intstr.FromString("value")
	fmt.Printf("ios2: %s\n", ios2.String()) // value

	ios3 := intstr.Parse("100")
	fmt.Printf("ios3 as string: %s\n", ios3.String()) // 100
	fmt.Printf("ios3 as int: %d\n", ios3.IntValue())  // 100

	ios4 := intstr.Parse("value")
	fmt.Printf("ios4 as string: %s\n", ios4.String()) // value
	fmt.Printf("ios4 as int: %d\n", ios4.IntValue())  // 0

	fmt.Printf("ios4 or 'default': %s\n", intstr.ValueOrDefault(&ios4, intstr.Parse("default"))) // value
	fmt.Printf("nil or 'default': %s\n", intstr.ValueOrDefault(nil, intstr.Parse("default")))    // default

	ios5 := intstr.Parse("10%")
	scaled, _ := intstr.GetScaledValueFromIntOrPercent(&ios5, 5000, true)
	fmt.Printf("10%% of 5000: %d\n", scaled) // 500
}
