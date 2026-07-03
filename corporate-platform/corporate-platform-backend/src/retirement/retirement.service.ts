import { Injectable } from '@nestjs/common';
import { CarbonAssetService } from '../stellar/soroban/contracts/carbon-asset.service';
import { IdempotencyService } from '../stellar/soroban/idempotency/idempotency.service';
import { DuplicateStrategy } from '../stellar/soroban/interfaces/idempotency.interface';

@Injectable()
export class RetirementService {
  constructor(
    private readonly carbonAssetService: CarbonAssetService,
    private readonly idempotencyService: IdempotencyService,
  ) {}

  async retireCredits(
    companyId: string,
    userId: string,
    creditId: string,
    amount: number,
    purpose: string,
  ): Promise<any> {
    const workflowId = `retirement-${creditId}-${Date.now()}`;

    // Check if already processed
    const isProcessed = await this.idempotencyService.isWorkflowProcessed(
      companyId,
      workflowId,
      'retire',
    );

    if (isProcessed) {
      // Return cached result
      const call = await this.idempotencyService.getContractCallByWorkflow(
        companyId,
        workflowId,
        'retire',
      );
      return {
        success: true,
        cached: true,
        transactionHash: call.transactionHash,
        result: call.result,
      };
    }

    // Execute with idempotency
    const result = await this.carbonAssetService.invokeWithIdempotency(
      companyId,
      {
        methodName: 'retire',
        args: [
          { type: 'u32', value: creditId },
          { type: 'u32', value: amount },
        ],
      },
      {
        workflowId,
        metadata: { userId, purpose },
      },
      DuplicateStrategy.RETURN_CACHED,
    );

    return {
      success: true,
      cached: result.isCached,
      isDuplicate: result.isDuplicate,
      transactionHash: result.transactionHash,
      callId: result.callId,
      workflowId: result.workflowId,
      result: result.result,
    };
  }
}