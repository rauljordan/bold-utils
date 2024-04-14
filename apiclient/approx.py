import numpy as np

YEAR = 365 * 24 * 60 * 60
BASIS_POINT_SCALE = 10000
EXP_SCALE = 1e18 

def exp_approx(rate):
    return EXP_SCALE - rate + rate**2 / (2 * EXP_SCALE) - rate**3 / (6 * EXP_SCALE**2)

# Our approximate implementation
def approximated(initial_balance, interest_rate_basis_points, t0, t1):
    time_diff = t1 - t0
    rate = (interest_rate_basis_points * time_diff * EXP_SCALE) / (BASIS_POINT_SCALE * YEAR)
    exp_approximation = exp_approx(rate)
    adjusted_balance = initial_balance * exp_approximation / EXP_SCALE
    return adjusted_balance

# Real implementation using numpy's exponential function
def real(initial_balance, interest_rate_basis_points, t0, t1):
    time_diff = t1 - t0
    rate = (interest_rate_basis_points * time_diff) / (BASIS_POINT_SCALE * YEAR)
    exp_rate = np.exp(-rate)
    adjusted_balance = initial_balance * exp_rate
    return adjusted_balance

initial_balance = 1000  # Initial balance
interest_rate_basis_points = 500  # Interest rate in basis points (5%)
t0 = 0  # Starting time
t1 = YEAR  # Ending time


# Calculating the error between our implementation and numpy's implementation
accrued_balance = approximated(initial_balance, interest_rate_basis_points, t0, t1)
accrued_balance_real = real(initial_balance, interest_rate_basis_points, t0, t1)

# Absolute error term
error = abs(accrued_balance - accrued_balance_real)
print(accrued_balance_real, accrued_balance, error)